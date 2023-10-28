package metrics 

import (
	"time"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/azukaar/cosmos-server/src/utils"
)

type DataDef struct {
	Max uint64 
	Period time.Duration
	Label string
	AggloType string
	SetOperation string
	Scale int
}

type DataPush struct {
	Date time.Time
	Key string
	Value int
	Max uint64
	Expire time.Time
	Label string
	AvgIndex int
	AggloType string
	Scale int
}

var dataBuffer = map[string]DataPush{}

var lock = make(chan bool, 1)

func MergeMetric(SetOperation string, currentValue int, newValue int, avgIndex int) int {
	if SetOperation == "" {  
		return newValue    
	} else if SetOperation == "max" {
		if newValue > currentValue {
			return newValue
		} else {
			return currentValue
		}
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
	nbData := 0

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
	if errCo != nil {
			utils.Error("Metrics - Database Connect", errCo)
			return
	}

	lock <- true
	defer func() { <-lock }()

	for dpkey, dp := range dataBuffer {
		// utils.Debug("Metrics - Saving " + dp.Key + " " + strconv.Itoa(dp.Value) + " " + dp.Date.String() + " " + dp.Expire.String())
		
		if dp.Expire.Before(time.Now()) {
			delete(dataBuffer, dpkey)
			nbData++

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
					},
			}
			
			// This ensures that if the document doesn't exist, it'll be created
			options := options.Update().SetUpsert(true)

			_, err := c.UpdateOne(nil, filter, update, options)

			if err != nil {
				utils.Error("Metrics - Database Insert", err)
				return
			}
		}
	}

	utils.Debug("Data - Saved " + strconv.Itoa(nbData) + " entries")
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

func PushSetMetric(key string, value int, def DataDef) {
	go func() {
		key = "cosmos." + key
		date := ModuloTime(time.Now(), def.Period)
		cacheKey := key + date.String()

		lock <- true
		defer func() { <-lock }()

		if dp, ok := dataBuffer[cacheKey]; ok {
			dp.Max = def.Max
			dp.Value = MergeMetric(def.SetOperation, dp.Value, value, dp.AvgIndex)    
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
			}
		}
	}()
}

func Run() {
	utils.Debug("Metrics - Run")
	
	nextTime := ModuloTime(time.Now().Add(time.Second*30), time.Second*30)
	nextTime = nextTime.Add(time.Second * 2)

	utils.Debug("Metrics - Next run at " + nextTime.String())
	time.AfterFunc(nextTime.Sub(time.Now()), func() {
		go func() {
			GetSystemMetrics()
			SaveMetrics()
		}()

		Run()
	})
}

func Init() {
	InitAggl()
	GetSystemMetrics()
	Run()
}