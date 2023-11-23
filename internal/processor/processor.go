package processor

import (
	"fmt"
	"strconv"

	"github.com/LogBustersHackathon/backend/model"
	"github.com/LogBustersHackathon/backend/utils"
)

var whitelist = []int{7014, 8193, 8192}

func Process(closingChn chan struct{}, processChn chan []model.KafkaAlarm, publishChn chan model.AlarmResponse) (err error) {
	fmt.Printf("Processing messages...\n")

taskLoop:
	for {
		select {
		case <-closingChn:
			break taskLoop
		case alarms := <-processChn:
		alarmLoop:
			for _, alarm := range alarms {
				if !utils.Contains(whitelist, int(alarm.DeviceID)) {
					continue alarmLoop
				}

				data := model.AlarmResponse{}
				data.AlarmType = model.Memory
				data.Owner = strconv.FormatInt(alarm.DeviceID, 10)
				data.Status = model.Critical
				data.Value = strconv.FormatInt(alarm.Int2, 10)

				fmt.Printf("Processing new alarm...\n")

				publishChn <- data
			}
		}
	}

	return nil
}
