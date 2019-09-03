package types

import (
	"fmt"
	btypes "github.com/QOSGroup/qbase/types"
	"github.com/QOSGroup/qos/module/params"
	"github.com/QOSGroup/qos/types"
	qtypes "github.com/QOSGroup/qos/types"
	"strconv"
	"time"
)

var (
	defaultMaxValidatorCnt             = int64(21)
	defaultValidatorVotingStatusLen    = int64(10000)
	defaultValidatorVotingStatusLeast  = int64(7000)
	defaultValidatorSurvivalSecs       = int64(8 * 60 * 60)
	defaultDelegatorUnbondReturnHeight = int64(15 * 24 * 60 * 60 / qtypes.DefaultBlockInterval) // 1 day
	defaultDelegatorRedelegationHeight = int64(24 * 60 * 60 / qtypes.DefaultBlockInterval)      // 15 days
	defaultMaxEvidenceAge              = time.Duration(1814400000000000)                        // ~= 21 days
	defaultSlashFractionDoubleSign     = types.NewDecWithPrec(2, 1)                             // 0.2
	defaultSlashFractionDowntime       = types.NewDecWithPrec(1, 4)                             // 0.0001
)

var (
	ParamSpace = "stake"

	// keys for stake p
	KeyMaxValidatorCnt                   = []byte("max_validator_cnt")
	KeyValidatorVotingStatusLen          = []byte("voting_status_len")
	KeyValidatorVotingStatusLeast        = []byte("voting_status_least")
	KeyValidatorSurvivalSecs             = []byte("survival_secs")
	KeyDelegatorUnbondFrozenHeight       = []byte("unbond_frozen_height")
	KeyDelegatorRedelegationActiveHeight = []byte("redelegation_active_height")
	KeyMaxEvidenceAge                    = []byte("max_evidence_age")
	KeySlashFractionDoubleSign           = []byte("slash_fraction_double_sign")
	KeySlashFractionDowntime             = []byte("slash_fraction_downtime")
)

type Params struct {
	MaxValidatorCnt                   int64         `json:"max_validator_cnt"`
	ValidatorVotingStatusLen          int64         `json:"voting_status_len"`
	ValidatorVotingStatusLeast        int64         `json:"voting_status_least"`
	ValidatorSurvivalSecs             int64         `json:"survival_secs"`
	DelegatorUnbondFrozenHeight       int64         `json:"unbond_frozen_height"`
	DelegatorRedelegationActiveHeight int64         `json:"redelegation_active_height"`
	MaxEvidenceAge                    time.Duration `json:"max_evidence_age"`
	SlashFractionDoubleSign           types.Dec     `json:"slash_fraction_double_sign"`
	SlashFractionDowntime             types.Dec     `json:"slash_fraction_downtime"`
}

func (p *Params) KeyValuePairs() qtypes.KeyValuePairs {
	return qtypes.KeyValuePairs{
		{KeyMaxValidatorCnt, &p.MaxValidatorCnt},
		{KeyValidatorVotingStatusLen, &p.ValidatorVotingStatusLen},
		{KeyValidatorVotingStatusLeast, &p.ValidatorVotingStatusLeast},
		{KeyValidatorSurvivalSecs, &p.ValidatorSurvivalSecs},
		{KeyDelegatorUnbondFrozenHeight, &p.DelegatorUnbondFrozenHeight},
		{KeyDelegatorRedelegationActiveHeight, &p.DelegatorRedelegationActiveHeight},
		{KeyMaxEvidenceAge, &p.MaxEvidenceAge},
		{KeySlashFractionDoubleSign, &p.SlashFractionDoubleSign},
		{KeySlashFractionDowntime, &p.SlashFractionDowntime},
	}
}

func (p *Params) Validate(key string, value string) (interface{}, btypes.Error) {
	switch key {
	case string(KeyMaxValidatorCnt),
		string(KeyValidatorVotingStatusLen),
		string(KeyValidatorVotingStatusLeast),
		string(KeyValidatorSurvivalSecs),
		string(KeyDelegatorUnbondFrozenHeight),
		string(KeyDelegatorRedelegationActiveHeight),
		string(KeyMaxEvidenceAge):
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil || v <= 0 {
			return nil, params.ErrInvalidParam(fmt.Sprintf("%s invalid", key))
		}
		return v, nil
	default:
		return nil, params.ErrInvalidParam(fmt.Sprintf("%s not exists", key))
	}
}

func (p *Params) GetParamSpace() string {
	return ParamSpace
}

func NewParams(maxValidatorCnt, validatorVotingStatusLen, validatorVotingStatusLeast, validatorSurvivalSecs,
	delegatorUnbondFrozenHeight, delegatorRedelegationActiveHeight int64, maxEvidenceAge time.Duration,
	slashFractionDoubleSign types.Dec, slashFractionDowntime types.Dec) Params {

	return Params{
		MaxValidatorCnt:                   maxValidatorCnt,
		ValidatorVotingStatusLen:          validatorVotingStatusLen,
		ValidatorVotingStatusLeast:        validatorVotingStatusLeast,
		ValidatorSurvivalSecs:             validatorSurvivalSecs,
		DelegatorUnbondFrozenHeight:       delegatorUnbondFrozenHeight,
		DelegatorRedelegationActiveHeight: delegatorRedelegationActiveHeight,
		MaxEvidenceAge:                    maxEvidenceAge,
		SlashFractionDoubleSign:           slashFractionDoubleSign,
		SlashFractionDowntime:             slashFractionDowntime,
	}
}

func DefaultParams() Params {
	return NewParams(defaultMaxValidatorCnt, defaultValidatorVotingStatusLen, defaultValidatorVotingStatusLeast, defaultValidatorSurvivalSecs, defaultDelegatorUnbondReturnHeight, defaultDelegatorRedelegationHeight, defaultMaxEvidenceAge, defaultSlashFractionDoubleSign, defaultSlashFractionDowntime)
}
