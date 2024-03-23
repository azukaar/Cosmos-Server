package metrics 

import (
	"time"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	
	"github.com/azukaar/cosmos-server/src/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

type DataDef struct {
	Max uint64 
	Period time.Duration
	Label string
	AggloType string
	SetOperation string
	Scale int
	Unit string
	Decumulate bool
	DecumulatePos bool
	Object string
}

type DataPush struct {
	Date time.Time
	Key string
	Value int
	Max uint64
	Period time.Duration
	Expire time.Time
	Label string
	AvgIndex int
	AggloType string
	Scale int
	Unit string
	Decumulate bool
	DecumulatePos bool
	Object string
}

var dataBuffer = map[string]DataPush{}

var lock = make(chan bool, 1)

func GetDataBuffer() map[string]DataPush {
	return dataBuffer
}

func MergeMetric(SetOperation string, currentValue int, newValue int, avgIndex int) int {
	if SetOperation == "" {  
		return newValue    
	} else if SetOperation == "max" {
		if newValue > currentValue {
			return newValue
		} else {
			return currentValue
		}
	} else if SetOperation == "sum" {
		return currentValue + newValue
	} else if SetOperation == "min" {
		if newValue < currentValue {
			return newValue
		} else {
			return currentValue
		}
	} else if SetOperation == "avg" {
		if avgIndex == 0 {
			avgIndex = 1
			return newValue
		} else {
			return (currentValue * (avgIndex) + newValue) / (avgIndex + 1)
		}
	} else {
		return newValue
	}
}

func SaveMetrics() {
	utils.Debug("Metrics - Saving data")
	utils.Debug("Time: " + time.Now().String())

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
	if errCo != nil {
		utils.Error("Metrics - Database Connect", errCo)
		return
	}

	lock <- true
	defer func() { <-lock }()

	var operations []mongo.WriteModel

	for dpkey, dp := range dataBuffer {
		if dp.Expire.Before(time.Now()) {
			delete(dataBuffer, dpkey)

			scale := 1
			if dp.Scale != 0 {
				scale = dp.Scale
			}

			filter := bson.M{"Key": dp.Key}
			update := bson.M{
				"$push": bson.M{"Values": 
					bson.M{
						"Date": dp.Date,
						"Value": dp.Value,
					},
				},
				"$set": bson.M{
					"LastUpdate": dp.Date,
					"Max": dp.Max,
					"Label": dp.Label,
					"AggloType": dp.AggloType,
					"Scale": scale,
					"Unit": dp.Unit,
					"Object": dp.Object,
					"TimeScale": float64(dp.Period / (time.Second * 30)),
				},
			}

			CheckAlerts(dp.Key, "latest", utils.AlertMetricTrack{
				Key: dp.Key,
				Object: dp.Object,
				Max: dp.Max,
			}, dp.Value)
			
			// Create a new UpdateOneModel
			operation := mongo.NewUpdateOneModel()
			operation.SetFilter(filter)
			operation.SetUpdate(update)
			operation.SetUpsert(true)
			
			// Append to operations
			operations = append(operations, operation)
		}
	}

	if len(operations) > 0 {
		_, err := c.BulkWrite(nil, operations)
		if err != nil {
			utils.Error("Metrics - Bulk Write Error", err)
		} else {
			utils.Debug("Data - Bulk Saved " + strconv.Itoa(len(operations)) + " entries")
		}
	} else {
		utils.Debug("No data to save")
	}
}

func ModuloTime(start time.Time, modulo time.Duration) time.Time {
	elapsed := start.UnixNano() // This gives us the total nanoseconds since 1970
	durationNano := modulo.Nanoseconds()

	// Here we take modulo of elapsed time with the duration to get the remainder.
	// Then, we subtract the remainder from the elapsed time to get the start of the period.
	roundedElapsed := elapsed - elapsed%durationNano

	// Convert back to time.Time
	return time.Unix(0, roundedElapsed)
}

var lastInserted = map[string]int{}

func PushSetMetric(key string, value int, def DataDef) {
	go func() {
		originalValue := value
		key = "cosmos." + key
		date := ModuloTime(time.Now(), def.Period)
		cacheKey := key + date.String()

		lock <- true
		defer func() { <-lock }()

		if def.Decumulate || def.DecumulatePos {
			if lastInserted[key] != 0 {
				value = value - lastInserted[key]
				if def.DecumulatePos && value < 0 {
					value = 0
				}
			} else {
				value = 0
			}
		}
		


		if dp, ok := dataBuffer[cacheKey]; ok {
			value = MergeMetric(def.SetOperation, dp.Value, value, dp.AvgIndex)    

			dp.Max = def.Max
			dp.Value = value
			if def.SetOperation == "avg" {
				dp.AvgIndex++
			}

			dataBuffer[cacheKey] = dp
		} else {
			dataBuffer[cacheKey] = DataPush{
				Date:   date,
				Expire: ModuloTime(time.Now().Add(def.Period), def.Period),
				Key:    key,
				Value:  value,
				Max:    def.Max,
				Label:  def.Label,
				AggloType: def.AggloType,
				Scale: def.Scale,
				Unit: def.Unit,
				Object: def.Object,
				Period: def.Period,
			}
		}

		lastInserted[key] = originalValue
	}()
}

func Run() {
	nextTime := ModuloTime(time.Now().Add(time.Second*30), time.Second*30)
	nextTime = nextTime.Add(time.Second * 2)

	
	if utils.GetMainConfig().MonitoringDisabled {
		time.AfterFunc(nextTime.Sub(time.Now()), func() {
			Run()
		})
		
		return
	}
	
	utils.Debug("Metrics - Run - Next run at " + nextTime.String())

	time.AfterFunc(nextTime.Sub(time.Now()), func() {
		go func() {
			GetSystemMetrics()
			SaveMetrics()
		}()

		Run()
	})
}

func Init() {
	lastInserted = map[string]int{}

	InitAggl()
	Run()

	if !utils.GetMainConfig().MonitoringDisabled {
		go GetSystemMetrics()
	}
}