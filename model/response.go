package model

type KafkaAlarm struct {
	String1   string `json:"string1"`
	String2   string `json:"string2"`
	String3   string `json:"string3"`
	String4   string `json:"string4"`
	String5   string `json:"string5"`
	String6   string `json:"string6"`
	String7   string `json:"string7"`
	DeviceID  int64  `json:"deviceid"`
	Epoch     int64  `json:"epoch"`
	Int1      int64  `json:"int1"`
	Int2      int64  `json:"int2"`
	Int3      int64  `json:"int3"`
	Int4      int64  `json:"int4"`
	Int5      int64  `json:"int5"`
	Int6      int64  `json:"int6"`
	Int7      int64  `json:"int7"`
	Retention int64  `json:"retention"`
	Type      int16  `json:"type"`
	InputType int16  `json:"inputtype"`
}

type AlarmResponse struct {
	Value     string
	Owner     string
	Status    StatusType
	AlarmType AlarmType
}

type StatusType int

const (
	Healthy  StatusType = 0
	Critical StatusType = 1
	Warning  StatusType = 2
)

type AlarmType int

const (
	Memory AlarmType = 1
)
