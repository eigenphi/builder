package miner

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"time"
)

// / To use it:
// / 1. Copy relevant data from the worker
// / 2. Call buildBlock
// / 2. If new bundles, txs arrive, call buildBlock again
// / This struct lifecycle is tied to 1 block-building task
type greedyBuilder struct {
	inputEnvironment *environment
	chainData        chainData
	interrupt        *int32
}

func newGreedyBuilder(chain *core.BlockChain, chainConfig *params.ChainConfig, blacklist map[common.Address]struct{}, env *environment, interrupt *int32) *greedyBuilder {
	return &greedyBuilder{
		inputEnvironment: env,
		chainData:        chainData{chainConfig, chain, blacklist},
		interrupt:        interrupt,
	}
}

func (b *greedyBuilder) mergeOrdersIntoEnvDiff(envDiff *environmentDiff, orders *types.TransactionsByPriceAndNonce) []types.SimulatedBundle {
	usedBundles := []types.SimulatedBundle{}

	txCount := 0
	bundleCount := 0
	start := time.Now()
	for {
		order := orders.Peek()
		if order == nil {
			break
		}

		if tx := order.Tx(); tx != nil {
			txCount += 1
			receipt, skip, err := envDiff.commitTx(tx, b.chainData)
			switch skip {
			case shiftTx:
				orders.Shift()
			case popTx:
				orders.Pop()
			}

			if err != nil {
				log.Trace("could not apply tx", "hash", tx.Hash(), "err", err)
				continue
			}
			effGapPrice, err := tx.EffectiveGasTip(envDiff.baseEnvironment.header.BaseFee)
			if err == nil {
				log.Info("Included tx", "EGP", effGapPrice.String(), "gasUsed", receipt.GasUsed,
					"txHash", tx.Hash().Hex())
			}
		} else if bundle := order.Bundle(); bundle != nil {
			bundleCount += 1
			//log.Debug("buildBlock considering bundle", "egp", bundle.MevGasPrice.String(), "hash", bundle.OriginalBundle.Hash)
			err := envDiff.commitBundle(bundle, b.chainData, b.interrupt)
			orders.Pop()
			if err != nil {
				log.Trace("Could not apply bundle", "bundle", bundle.OriginalBundle.Hash, "err", err)
				continue
			}

			log.Info("Included bundle", "bundleEGP", bundle.MevGasPrice.String(),
				"gasUsed", bundle.TotalGasUsed,
				"ethToCoinbase", ethIntToFloat(bundle.TotalEth),
				"hash", bundle.OriginalBundle.Hash.Hex())
			usedBundles = append(usedBundles, *bundle)
		}
	}
	log.Info("ApplyTxAndBundles", "txCount", txCount, "bundleCount", bundleCount, "time", time.Since(start))

	return usedBundles
}

func (b *greedyBuilder) buildBlock(simBundles []types.SimulatedBundle, transactions map[common.Address]types.Transactions) (*environment, []types.SimulatedBundle) {
	orders := types.NewTransactionsByPriceAndNonce(b.inputEnvironment.signer, transactions, simBundles, b.inputEnvironment.header.BaseFee)
	envDiff := newEnvironmentDiff(b.inputEnvironment.copy())
	usedBundles := b.mergeOrdersIntoEnvDiff(envDiff, orders)
	envDiff.applyToBaseEnv()
	return envDiff.baseEnvironment, usedBundles
}
