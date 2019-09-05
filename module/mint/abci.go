package mint

import (
	"github.com/QOSGroup/qbase/context"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/QOSGroup/qos/module/distribution"
	"github.com/QOSGroup/qos/module/mint/mapper"
	"github.com/QOSGroup/qos/module/mint/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

//BeginBlocker: 挖矿奖励
func BeginBlocker(ctx context.Context, req abci.RequestBeginBlock) {
	height := ctx.BlockHeight()
	currentBlockTime := ctx.BlockHeader().Time.UTC()

	mintMapper := mapper.GetMapper(ctx)
	distributionMapper := distribution.GetMapper(ctx)

	phrases := mintMapper.MustGetInflationPhrases()

	// 当前通胀校验
	currentPhrase, exists := phrases.GetPhrase(currentBlockTime)
	if !exists || !currentPhrase.TotalAmount.GT(btypes.ZeroInt()) || currentPhrase.AppliedAmount.Equal(currentPhrase.TotalAmount) {
		return
	}

	// 处理前一通胀阶段未完整发行情况，剩余转到社区账户
	if currentPhrase.AppliedAmount.Equal(btypes.ZeroInt()) {
		if prePhrase, exists := phrases.GetPrePhrase(currentBlockTime); exists {
			if !prePhrase.AppliedAmount.Equal(prePhrase.TotalAmount) {
				prePhraseLeft := prePhrase.TotalAmount.Sub(prePhrase.AppliedAmount)
				phrases = phrases.ApplyQOS(prePhrase.EndTime, prePhraseLeft)
				distributionMapper.AddToCommunityFeePool(prePhraseLeft)
			}
		}

	}

	if height == 1 {
		mintMapper.SetFirstBlockTime(currentBlockTime.Unix())
	} else {
		// 计算出快时间
		firstBlockTime := mintMapper.GetFirstBlockTime()
		blockTimeAvg := (currentBlockTime.Unix() - firstBlockTime) / (height - 1)

		// 计算挖矿奖励
		blocks := int64(currentPhrase.EndTime.Sub(currentBlockTime).Seconds()) / blockTimeAvg
		rewardPerBlock := currentPhrase.TotalAmount.Sub(currentPhrase.AppliedAmount).DivRaw(blocks)

		if rewardPerBlock.GT(btypes.ZeroInt()) {
			// 保存通胀发行更新
			mintMapper.AddAllTotalMintQOSAmount(rewardPerBlock)
			phrases := phrases.ApplyQOS(currentPhrase.EndTime, rewardPerBlock)
			mintMapper.SetInflationPhrases(phrases)

			// 挖矿奖励保存至待分配
			distributionMapper.AddPreDistributionQOS(rewardPerBlock)

			ctx.EventManager().EmitEvent(
				btypes.NewEvent(
					types.EventTypeMint,
					btypes.NewAttribute(types.AttributeKeyHeight, string(height)),
					btypes.NewAttribute(types.AttributeKeyTokens, rewardPerBlock.String()),
				),
			)
		}
	}
}
