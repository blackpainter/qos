package staking

import (
	"github.com/QOSGroup/qbase/baseabci"
	"github.com/QOSGroup/qbase/context"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/QOSGroup/qos/mapper"
	"github.com/QOSGroup/qos/types"

	qacc "github.com/QOSGroup/qos/account"
	abci "github.com/tendermint/tendermint/abci/types"
)

//1. 统计validator投票信息, 将不活跃的validator转成InActive状态
func BeginBlocker(ctx context.Context, req abci.RequestBeginBlock) {

	mainMapper := mapper.GetMainMapper(ctx)
	validatorMapper := GetValidatorMapper(ctx)

	votingWindowLen := uint64(mainMapper.GetStakeConfig().ValidatorVotingStatusLen)
	minVotingCounter := uint64(mainMapper.GetStakeConfig().ValidatorVotingStatusLeast)

	lastValidators := make([]abci.Validator, 0, len(req.LastCommitInfo.Validators))

	for _, signingValidator := range req.LastCommitInfo.Validators {
		valAddr := btypes.Address(signingValidator.Validator.Address)
		lastValidators = append(lastValidators, signingValidator.Validator)
		voted := signingValidator.SignedLastBlock
		handleValidatorValidatorVoteInfo(ctx, valAddr, voted, votingWindowLen, minVotingCounter)
	}

	//保存上一次validator 地址集合
	validatorMapper.Set(BuildLastValidatorAddressSetKey(), lastValidators)
}

//1. 将所有InActive到一定期限的validator删除
//2. 统计新的validator
func EndBlocker(ctx context.Context) (res abci.ResponseEndBlock) {

	mainMapper := mapper.GetMainMapper(ctx)
	survivalSecs := mainMapper.GetStakeConfig().ValidatorSurvivalSecs
	maxValidatorCount := uint64(mainMapper.GetStakeConfig().MaxValidatorCnt)

	closeExpireInactiveValidator(ctx, survivalSecs)
	res.ValidatorUpdates = getUpdateValidators(ctx, maxValidatorCount)
	return
}

func closeExpireInactiveValidator(ctx context.Context, survivalSecs uint64) {
	log := ctx.Logger()
	validatorMapper := GetValidatorMapper(ctx)
	voteInfoMapper := GetVoteInfoMapper(ctx)
	accountMapper := baseabci.GetAccountMapper(ctx)

	blockTimeSec := uint64(ctx.BlockHeader().Time.UTC().Unix())
	lastCloseValidatorSec := blockTimeSec - survivalSecs

	iterator := validatorMapper.IteratorInActiveValidator(uint64(0), lastCloseValidatorSec)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		valAddress := btypes.Address(key[9:])

		log.Info("close validator", "height", ctx.BlockHeight(), "validator", valAddress.String())
		if validator, ok := validatorMapper.KickValidator(valAddress); ok {

			voteInfoMapper.DelValidatorVoteInfo(valAddress)
			voteInfoMapper.ClearValidatorVoteInfoInWindow(valAddress)

			//关闭validator后,归还绑定的token
			owner := accountMapper.GetAccount(validator.Owner)
			if qosAcc, ok := owner.(*qacc.QOSAccount); ok {
				backQOS := btypes.NewInt(int64(validator.BondTokens))
				qosAcc.SetQOS(qosAcc.GetQOS().NilToZero().Add(backQOS))
				accountMapper.SetAccount(qosAcc)
			}
		}
	}
}

func getUpdateValidators(ctx context.Context, maxValidatorCount uint64) []abci.Validator {
	log := ctx.Logger()
	validatorMapper := ctx.Mapper(ValidatorMapperName).(*ValidatorMapper)

	var lastValidators []abci.Validator
	validatorMapper.Get(BuildLastValidatorAddressSetKey(), &lastValidators)

	lastValidatorMap := make(map[string]abci.Validator)
	for _, lastValidator := range lastValidators {
		lastValidatorAddr := btypes.Address(lastValidator.Address).String()
		lastValidatorMap[lastValidatorAddr] = lastValidator
	}

	updateValidators := make([]abci.Validator, 0, len(lastValidatorMap))

	i := uint64(0)
	validatorsMap := make(map[string]abci.Validator)

	iterator := validatorMapper.IteratorValidatrorByVoterPower(false)
	defer iterator.Close()

	var key []byte
	for ; iterator.Valid(); iterator.Next() {
		key = iterator.Key()
		valAddr := btypes.Address(key[9:])

		if i >= maxValidatorCount {
			//超出MaxValidatorCnt的validator修改为Inactive状态
			validatorMapper.MakeValidatorInActive(valAddr, uint64(ctx.BlockHeight()), ctx.BlockHeader().Time, types.MaxValidator)
		} else {
			if validator, exsits := validatorMapper.GetValidator(valAddr); exsits {
				i++
				abciValidator := validator.ToABCIValidator()
				abciAddr := btypes.Address(abciValidator.Address).String()
				validatorsMap[abciAddr] = abciValidator

				//新增或修改
				lastValidator, exsits := lastValidatorMap[abciAddr]
				if !exsits || (abciValidator.Power != lastValidator.Power) {
					updateValidators = append(updateValidators, abciValidator)
				}
			}
		}
	}

	//删除
	for lastValidatorAddr, lastValidator := range lastValidatorMap {
		if _, ok := validatorsMap[lastValidatorAddr]; !ok {
			lastValidator.Power = 0
			updateValidators = append(updateValidators, lastValidator)
		}
	}

	log.Info("update Validators", "len", len(updateValidators))

	return updateValidators
}

func handleValidatorValidatorVoteInfo(ctx context.Context, valAddr btypes.Address, isVote bool, votingWindowLen, minVotingCounter uint64) {

	log := ctx.Logger()
	height := uint64(ctx.BlockHeight())
	validatorMapper := GetValidatorMapper(ctx)
	voteInfoMapper := GetVoteInfoMapper(ctx)

	validator, exsits := validatorMapper.GetValidator(valAddr)
	if !exsits {
		log.Info("validatorVoteInfo", valAddr.String(), "not exsits,may be closed")
		return
	}

	//非Active状态不处理
	if !validator.IsActive() {
		log.Info("validatorVoteInfo", valAddr.String(), "is InActive")
		return
	}

	voteInfo, exsits := voteInfoMapper.GetValidatorVoteInfo(valAddr)
	if !exsits {
		voteInfo = types.NewValidatorVoteInfo(height, 0, 0)
	}

	index := voteInfo.IndexOffset % votingWindowLen
	voteInfo.IndexOffset++

	previousVote := voteInfoMapper.GetVoteInfoInWindow(valAddr, index)

	switch {
	case previousVote && !isVote:
		voteInfoMapper.SetVoteInfoInWindow(valAddr, index, false)
		voteInfo.MissedBlocksCounter++
	case !previousVote && isVote:
		voteInfoMapper.SetVoteInfoInWindow(valAddr, index, true)
		voteInfo.MissedBlocksCounter--
	default:
		//nothing
	}

	if !isVote {
		log.Info("validatorVoteInfo", "height", height, valAddr.String(), "not vote")
	}

	// minHeight := voteInfo.StartHeight + votingWindowLen
	maxMissedCounter := votingWindowLen - minVotingCounter

	// if height > minHeight && voteInfo.MissedBlocksCounter > maxMissedCounter
	if voteInfo.MissedBlocksCounter > maxMissedCounter {
		log.Info("validator to inActive", "height", height, "validator", valAddr.String(), "missed counter", voteInfo.MissedBlocksCounter)

		blockValidator(ctx, valAddr)

		// voteInfo.IndexOffset = 0
		// voteInfo.MissedBlocksCounter = 0
		// voteInfoMapper.ClearValidatorVoteInfoInWindow(valAddr)
	}

	voteInfoMapper.SetValidatorVoteInfo(valAddr, voteInfo)
}

//
func blockValidator(ctx context.Context, valAddr btypes.Address) {
	validatorMapper := GetValidatorMapper(ctx)
	validatorMapper.MakeValidatorInActive(valAddr, uint64(ctx.BlockHeight()), ctx.BlockHeader().Time, types.MissVoteBlock)
}