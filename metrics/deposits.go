package metrics

import (
	//"github.com/alrevuelta/eth-pools-metrics/prometheus"
	log "github.com/sirupsen/logrus"
	"runtime"
	"time"
)

// TODO: Temporal solution:
// - TheGraph API calls has some limits, so we can't query in every epoch
// - Race condition with the depositedKeys
// - Fetches the deposits every hour
func (a *Metrics) StreamDeposits() {
	for {
		pubKeysDeposited, err := a.theGraph.GetAllDepositedKeys()
		if err != nil {
			log.Error(err)
			time.Sleep(10 * 60 * time.Second)
			continue
		}
		a.depositedKeys = pubKeysDeposited

		log.WithFields(log.Fields{
			"DepositedValidators": len(pubKeysDeposited),
			// TODO: Print epoch
			//"Slot":     slot,
			//"Epoch":    uint64(slot) % a.slotsInEpoch,
		}).Info("Deposits:")

		// Temporal fix to memory leak. Perhaps having an infinite loop
		// inside a routinne is not a good idea. TODO
		runtime.GC()

		time.Sleep(60 * 60 * time.Second)
	}
}
