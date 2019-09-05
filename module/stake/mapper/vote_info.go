package mapper

import (
	"github.com/QOSGroup/qos/module/stake/types"

	btypes "github.com/QOSGroup/qbase/types"
)

func (mapper *Mapper) GetValidatorVoteInfo(valAddr btypes.ValAddress) (info types.ValidatorVoteInfo, exists bool) {
	key := types.BuildValidatorVoteInfoKey(valAddr)
	exists = mapper.Get(key, &info)
	return
}

func (mapper *Mapper) SetValidatorVoteInfo(valAddr btypes.ValAddress, info types.ValidatorVoteInfo) {
	key := types.BuildValidatorVoteInfoKey(valAddr)
	mapper.Set(key, info)
}

func (mapper *Mapper) ResetValidatorVoteInfo(valAddr btypes.ValAddress, info types.ValidatorVoteInfo) {
	key := types.BuildValidatorVoteInfoKey(valAddr)
	mapper.ClearValidatorVoteInfoInWindow(valAddr)
	mapper.Del(key)
}

func (mapper *Mapper) DelValidatorVoteInfo(valAddr btypes.ValAddress) {
	key := types.BuildValidatorVoteInfoKey(valAddr)
	mapper.Del(key)
}

func (mapper *Mapper) GetVoteInfoInWindow(valAddr btypes.ValAddress, index int64) (vote bool) {
	key := types.BuildValidatorVoteInfoInWindowKey(index, valAddr)
	vote, exists := mapper.GetBool(key)

	if !exists {
		return true
	}

	return vote
}

func (mapper *Mapper) SetVoteInfoInWindow(valAddr btypes.ValAddress, index int64, vote bool) {
	key := types.BuildValidatorVoteInfoInWindowKey(index, valAddr)
	mapper.Set(key, vote)
}

func (mapper *Mapper) ClearValidatorVoteInfoInWindow(valAddr btypes.ValAddress) {
	prefixKey := append(types.GetValidatorVoteInfoInWindowKey(), valAddr...)
	endKey := btypes.PrefixEndBytes(prefixKey)
	iter := mapper.GetStore().Iterator(prefixKey, endKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		mapper.Del(iter.Key())
	}
}

//-------------------------genesis export

func (mapper *Mapper) IterateVoteInfos(fn func(btypes.ValAddress, types.ValidatorVoteInfo)) {
	iter := btypes.KVStorePrefixIterator(mapper.GetStore(), types.GetValidatorVoteInfoKey())
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		valAddr := types.GetValidatorVoteInfoAddr(key)
		var info types.ValidatorVoteInfo
		mapper.DecodeObject(iter.Value(), &info)
		fn(valAddr, info)
	}
}

func (mapper *Mapper) IterateVoteInWindowsInfos(fn func(int64, btypes.ValAddress, bool)) {
	iter := btypes.KVStorePrefixIterator(mapper.GetStore(), types.GetValidatorVoteInfoInWindowKey())
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		index, addr := types.GetValidatorVoteInfoInWindowIndexAddr(key)
		var vote bool
		mapper.DecodeObject(iter.Value(), &vote)
		fn(index, addr, vote)
	}
}
