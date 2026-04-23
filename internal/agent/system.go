package agent

import (
	"context"
	"fmt"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func (a *Agent) CollectSystemBatch(ctx context.Context) ([]models.Metrics, error) {
	var batch []models.Metrics

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	total := float64(vm.Total)
	free := float64(vm.Free)

	batch = append(batch, models.Metrics{
		ID:    "TotalMemory",
		MType: models.Gauge,
		Value: &total,
	})

	batch = append(batch, models.Metrics{
		ID:    "FreeMemory",
		MType: models.Gauge,
		Value: &free,
	})

	cpuValues, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, err
	}

	for i, v := range cpuValues {
		val := v
		name := fmt.Sprintf("CPUutilization%d", i+1)

		batch = append(batch, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &val,
		})
	}

	return batch, nil
}
