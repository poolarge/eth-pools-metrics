package metrics

import (
	"context"
	"github.com/alrevuelta/eth-pools-metrics/prometheus"
	"github.com/alrevuelta/eth-pools-metrics/schemas"
	ethpb "github.com/prysmaticlabs/prysm/v2/proto/prysm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"runtime"
	"time"
)

// TODO: Handle race condition
func (a *Metrics) StreamValidatorStatus() {
	for {
		if a.depositedKeys == nil {
			log.Warn("No depositedKeys are available")
			time.Sleep(10 * time.Second)
			continue
		}

		// Get the status of all the validators
		valsStatus, err := a.validatorClient.MultipleValidatorStatus(
			context.Background(),
			&ethpb.MultipleValidatorStatusRequest{
				PublicKeys: a.depositedKeys,
			})

		if err != nil {
			log.Error(err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Get validators with active duties
		validatingKeys := filterValidatingValidators(valsStatus)
		a.validatingKeys = validatingKeys

		// For other status we just want the count
		metrics := getValidatorStatusMetrics(valsStatus)
		logValidatorStatus(metrics)
		setPrometheusValidatorStatus(metrics)

		// Temporal fix to memory leak. Perhaps having an infinite loop
		// inside a routinne is not a good idea. TODO
		runtime.GC()

		time.Sleep(6 * 60 * time.Second)
	}
}

func setPrometheusValidatorStatus(metrics schemas.ValidatorStatusMetrics) {
	prometheus.NOfValidatingValidators.Set(float64(metrics.Validating))
	prometheus.NOfUnkownValidators.Set(float64(metrics.Unknown))
	prometheus.NOfDepositedValidators.Set(float64(metrics.Deposited))
	prometheus.NOfPendingValidators.Set(float64(metrics.Pending))
	prometheus.NOfActiveValidators.Set(float64(metrics.Active))
	prometheus.NOfExitingValidators.Set(float64(metrics.Exiting))
	prometheus.NOfSlashingValidators.Set(float64(metrics.Slashing))
	prometheus.NOfExitedValidators.Set(float64(metrics.Exited))
	prometheus.NOfInvalidValidators.Set(float64(metrics.Invalid))
	prometheus.NOfPartiallyDepositedValidators.Set(float64(metrics.PartiallyDeposited))
}

func filterValidatingValidators(vals *ethpb.MultipleValidatorStatusResponse) [][]byte {
	activeKeys := make([][]byte, 0)
	for i := range vals.PublicKeys {
		if isKeyValidating(vals.Statuses[i].Status) {
			activeKeys = append(activeKeys, vals.PublicKeys[i])
		}
	}
	return activeKeys
}

// Active as in having to fulfill duties
func isKeyValidating(status ethpb.ValidatorStatus) bool {
	if status == ethpb.ValidatorStatus_ACTIVE ||
		status == ethpb.ValidatorStatus_EXITING ||
		status == ethpb.ValidatorStatus_SLASHING {
		return true
	}
	return false
}

func getValidatorStatusMetrics(
	statusResponse *ethpb.MultipleValidatorStatusResponse) schemas.ValidatorStatusMetrics {

	metrics := schemas.ValidatorStatusMetrics{}
	for i := range statusResponse.PublicKeys {
		status := statusResponse.Statuses[i].Status

		// Note that a validator can be validating and active/exiting
		if isKeyValidating(status) {
			metrics.Validating++
		}

		if status == ethpb.ValidatorStatus_UNKNOWN_STATUS {
			metrics.Unknown++
		} else if status == ethpb.ValidatorStatus_DEPOSITED {
			metrics.Deposited++
		} else if status == ethpb.ValidatorStatus_PENDING {
			metrics.Pending++
		} else if status == ethpb.ValidatorStatus_ACTIVE {
			metrics.Active++
		} else if status == ethpb.ValidatorStatus_EXITING {
			metrics.Exiting++
		} else if status == ethpb.ValidatorStatus_SLASHING {
			metrics.Slashing++
		} else if status == ethpb.ValidatorStatus_EXITED {
			metrics.Exited++
		} else if status == ethpb.ValidatorStatus_INVALID {
			metrics.Invalid++
		} else if status == ethpb.ValidatorStatus_PARTIALLY_DEPOSITED {
			metrics.PartiallyDeposited++
		} else {
			log.Warn("Unknown status: ", status)
		}
	}
	return metrics
}

func logValidatorStatus(metrics schemas.ValidatorStatusMetrics) {
	log.WithFields(log.Fields{
		"Validating":         metrics.Validating,
		"Unknown":            metrics.Unknown,
		"Deposited":          metrics.Deposited,
		"Pending":            metrics.Pending,
		"Active":             metrics.Active,
		"Exiting":            metrics.Exiting,
		"Slashing":           metrics.Slashing,
		"Exited":             metrics.Exited,
		"Invalid":            metrics.Invalid,
		"PartiallyDeposited": metrics.PartiallyDeposited,
	}).Info("Validator Status Count:")
}
