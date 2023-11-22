package processor

import (
	"strconv"

	"github.com/LogBustersHackathon/backend/model"
	"github.com/LogBustersHackathon/backend/utils"
)

var whitelist = []int{7014, 8193, 8192}

func Process(closingChn chan struct{}, processChn chan []model.KafkaAlarm, publishChn chan model.AlarmResponse) (err error) {
taskLoop:
	for {
		select {
		case <-closingChn:
			break taskLoop
		case alarms := <-processChn:
			// temp := make(map[int64][]model.KafkaAlarm)
		alarmLoop:
			for _, alarm := range alarms {
				if alarm.Int1 != 439 {
					continue alarmLoop
				}

				if !utils.Contains(whitelist, int(alarm.DeviceID)) {
					continue alarmLoop
				}

				data := model.AlarmResponse{}
				data.AlarmType = model.Memory
				data.Owner = strconv.FormatInt(alarm.DeviceID, 10)
				data.Status = model.Critical
				data.Value = strconv.FormatInt(alarm.Int2, 10)

				publishChn <- data

				// temp[alarm.DeviceID] = append(temp[alarm.DeviceID], alarm)
			}

			// TODO: Process alarm here and send to NATS

			// publishChn <- struct{}{}
		}
	}

	return nil
}
