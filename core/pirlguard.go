package core

import (
	"errors"
	"git.pirl.io/bitcoiin/go-bitcoiin/core/types"
	"git.pirl.io/bitcoiin/go-bitcoiin/log"
	"git.pirl.io/bitcoiin/go-bitcoiin/params"
	"sort"
)

var syncStatus bool
//var maxReorgValue = 5
//var maxChangedHashes = 3

func (bc *BlockChain) checkChainForAttack(blocks types.Blocks) error {
	// Copyright 2014 The go-ethereum Authors
	// Copyright 2018 Pirl Sprl
	// This file is part of the go-ethereum library modified with Pirl Security Protocol.
	//
	// The go-ethereum library is free software: you can redistribute it and/or modify
	// it under the terms of the GNU Lesser General Public License as published by
	// the Free Software Foundation, either version 3 of the License, or
	// (at your option) any later version.
	//
	// The go-ethereum library is distributed in the hope that it will be useful,
	// but WITHOUT ANY WARRANTY; without even the implied warranty of
	// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
	// GNU Lesser General Public License for more details.
	//
	// You should have received a copy of the GNU Lesser General Public License
	// along with the go-ethereum library. If not, see http://www.gnu.org/licenses/.
	// Package core implements the Ethereum consensus protocol modified with Pirl Security Protocol.

	err := errors.New("")
	err = nil
	timeMap := make(map[uint64]int64)
	tipOfTheMainChain := bc.CurrentBlock().NumberU64()



	if !syncStatus {
		if tipOfTheMainChain == blocks[0].NumberU64() - 1 {
			//fmt.Println("We are synced")
			syncStatus = true
		} else {
			//fmt.Println("Still syncing!")
			syncStatus = false
		}
	}
	//counter := 0

	if len(blocks) > 0 && bc.CurrentBlock().NumberU64() > uint64(params.PirlGuardActivationBlock) {
		//if syncStatus && len(blocks) < int(params.PirlGuardBlockLength) && len(blocks) > maxReorgValue {
		//	//fmt.Println("We are in the condition here to check smaller block sizes...")
		//	for _, b := range blocks {
		//		//fmt.Println("This is the tx hash from incoming block : ",b.NumberU64()," with hash : " , b.Header().Hash().String())
		//		block := bc.GetBlockByNumber(b.NumberU64())
		//		if block != nil {
		//			//fmt.Println("This is the tx hash from db block : ",block.NumberU64()," with hash : " , block.Header().Hash().String())
		//			if b.Header().Hash().String() != block.Header().Hash().String() {
		//				counter++
		//				//fmt.Println("block tx hashes dont match for block : ", block.NumberU64())
		//			}
		//		} else {
		//			fmt.Println("block not found in db : ", b.NumberU64())
		//		}
		//		fmt.Println("Matching blocks with changed hashes : ", counter)
		//	}
		//	if counter > maxChangedHashes {
		//		fmt.Println("big reorg detected")
		//		return ErrBigReorg
		//	}
		//}
		if syncStatus && len(blocks) >= int(params.PirlGuardBlockLength) {
			for _, b := range blocks {
				timeMap[b.NumberU64()] = calculatePenaltyTimeForBlock(tipOfTheMainChain, b.NumberU64())
			}
		}
	}

	p := make(PairList, len(timeMap))
	index := 0
	for k, v := range timeMap {
		p[index] = Pair {k, v}
		index++
	}
	sort.Sort(p)
	var penalty int64
	for _, v := range p {
		penalty += v.Value
	}

	multi := calculateMulti(bc.CurrentBlock().Difficulty().Uint64())
	penalty = penalty * int64(multi)

	if penalty < 0 {
		penalty = 0
	}
	//fmt.Println("Penalty value for the chain :", penalty)
	context := []interface{}{
		"synced", syncStatus, "number", tipOfTheMainChain, "incoming_number", blocks[0].NumberU64() - 1, "penalty", penalty ,"implementation", "The Pirl Team --> https://pirl.io",
	}

	log.Info("checking legitimity of the chain", context... )

	if penalty > 0 {
		context := []interface{}{
			"penalty", penalty,
		}
		log.Error("Chain is a malicious and we should reject it", context... )
		err = ErrDelayTooHigh

	}

	if penalty == 0 {
		err = nil
	}

	return err
}

func calculatePenaltyTimeForBlock(tipOfTheMainChain , incomingBlock uint64) int64 {
	if incomingBlock < tipOfTheMainChain {
		return int64(tipOfTheMainChain - incomingBlock)
	}
	if incomingBlock == tipOfTheMainChain {
		return 0
	}
	if incomingBlock > tipOfTheMainChain {
		return -1
	}
	return 0
}

func calculateMulti(diff uint64) uint64 {

	if diff <= 500000000 {
		return 5
	}
	if diff >= 500000000 && diff < 20000000000 {
		return 4
	}
	if diff >= 20000000000 && diff < 30000000000 {
		return 3
	}
	if diff >= 30000000000 && diff < 50000000000 {
		return 2
	}
	if diff >= 50000000000 {
		return 1
	}
	return 1
}

// A data structure to hold key/value pairs
type Pair struct {
	Key   uint64
	Value int64
}

// A slice of pairs that implements sort.Interface to sort by values
type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Key < p[j].Key }