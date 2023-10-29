package metrics 

import (
	"time"

	"github.com/jasonlvhit/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/azukaar/cosmos-server/src/utils"
)

type DataDefDBEntry struct {
	Date time.Time
	Value int
	Processed bool

	// For agglomeration
	AvgIndex int
	AggloTo time.Time
	AggloExpire time.Time
}

type DataDefDB struct {
	Values []DataDefDBEntry
	ValuesAggl map[string]DataDefDBEntry
	LastUpdate time.Time
	Max uint64
	Label string
	Key string
	AggloType string
	Scale int
}

func AggloMetrics() []DataDefDB {
	lock <- true
	defer func() { <-lock }()

	utils.Log("Metrics: Agglomeration started")

	utils.Debug("Time: " + time.Now().String())

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
	if errCo != nil {
			utils.Error("Metrics - Database Connect", errCo)
			return []DataDefDB{}
	}

	// get all metrics from database
	var metrics []DataDefDB
	cursor, err := c.Find(nil, map[string]interface{}{})
	if err != nil {
		utils.Error("Metrics: Error fetching metrics", err)
		return []DataDefDB{}
	}
	defer cursor.Close(nil)
	
	if err = cursor.All(nil, &metrics); err != nil {
		utils.Error("Metrics: Error decoding metrics", err)
		return []DataDefDB{}
	}

	// populate aggregation pools
	hourlyPool := ModuloTime(time.Now(), time.Hour)
	hourlyPoolTo := ModuloTime(time.Now().Add(1 * time.Hour), time.Hour)
	dailyPool := ModuloTime(time.Now(), 24 * time.Hour)
	dailyPoolTo := ModuloTime(time.Now().Add(24 * time.Hour), 24 * time.Hour)
	
	for metInd, metric := range metrics {
		values := metric.Values
		
		// init map
		if metric.ValuesAggl == nil {
			metric.ValuesAggl = map[string]DataDefDBEntry{}
		} 

		// if hourly pool does not exist, create it
		if _, ok := metric.ValuesAggl["hour_" + hourlyPool.UTC().Format("2006-01-02 15:04:05")]; !ok {
			metric.ValuesAggl["hour_" + hourlyPool.UTC().Format("2006-01-02 15:04:05")] = DataDefDBEntry{
				Date: hourlyPool,
				Value: 0,
				Processed: false,
				AvgIndex: 0,
				AggloTo: hourlyPoolTo,
				AggloExpire: hourlyPoolTo.Add(48 * time.Hour),
			}
		}
	
		// if daily pool does not exist, create it
		if _, ok := metric.ValuesAggl["day_" + dailyPool.UTC().Format("2006-01-02")]; !ok {
			metric.ValuesAggl["day_" + dailyPool.UTC().Format("2006-01-02")] = DataDefDBEntry{
				Date: dailyPool,
				Value: 0,
				Processed: false,
				AvgIndex: 0,
				AggloTo: dailyPoolTo,
				AggloExpire: dailyPoolTo.Add(30 * 24 * time.Hour),
			}
		}

		for valInd, value := range values {
			// if not processed
			if !value.Processed {
				valueHourlyPool := ModuloTime(value.Date, time.Hour).UTC().Format("2006-01-02 15:04:05")
				valueDailyPool := ModuloTime(value.Date, 24 * time.Hour).UTC().Format("2006-01-02")

				if _, ok := metric.ValuesAggl["hour_" + valueHourlyPool]; ok {
					currentPool := metric.ValuesAggl["hour_" + valueHourlyPool]
					
					currentPool.Value = MergeMetric(metric.AggloType, currentPool.Value, value.Value, currentPool.AvgIndex)    
					if metric.AggloType == "avg" {
						currentPool.AvgIndex++
					}

					metric.ValuesAggl["hour_" + valueHourlyPool] = currentPool
				} else {
					utils.Warn("Metrics: Agglomeration - Pool not found : " + "hour_" + valueHourlyPool)
				}

				if _, ok := metric.ValuesAggl["day_" + valueDailyPool]; ok {
					currentPool := metric.ValuesAggl["day_" + valueDailyPool]
					
					currentPool.Value = MergeMetric(metric.AggloType, currentPool.Value, value.Value, currentPool.AvgIndex)
					if metric.AggloType == "avg" {
						currentPool.AvgIndex++
					}

					metric.ValuesAggl["day_" + valueDailyPool] = currentPool
				} else {
					utils.Warn("Metrics: Agglomeration - Pool not found: " + "day_" + valueDailyPool)
				}

				values[valInd].Processed = true
			}
		}

		// delete values over 1h
		finalValues := []DataDefDBEntry{}
		for _, value := range values {
			if value.Date.After(time.Now().Add(-1 * time.Hour)) {
				finalValues = append(finalValues, value)
			}
		}

		metric.Values = finalValues

		// clean up old agglo values
		for aggloKey, aggloValue := range metric.ValuesAggl {
			if aggloValue.AggloExpire.Before(time.Now()) {
				delete(metric.ValuesAggl, aggloKey)
			}
		}

		metrics[metInd] = metric
	}

	return metrics
}

func CommitAggl(metrics []DataDefDB) {
	lock <- true
	defer func() { <-lock }()

	utils.Log("Metrics: Agglomeration done. Saving to DB")
	
	c, errCo := utils.GetCollection(utils.GetRootAppId(), "metrics")
	if errCo != nil {
			utils.Error("Metrics - Database Connect", errCo)
			return
	}

	// save metrics
	for _, metric := range metrics {
		options := options.Update().SetUpsert(true)

		_, err := c.UpdateOne(nil, bson.M{"Key": metric.Key}, bson.M{"$set": bson.M{"Values": metric.Values, "ValuesAggl": metric.ValuesAggl}}, options)

		if err != nil {
			utils.Error("Metrics: Error saving metrics", err)
			return
		}
	}

	utils.Log("Metrics: Agglomeration saved to DB")
}

func AggloAndCommitMetrics() {
	CommitAggl(AggloMetrics())
}

func InitAggl() {
	go func() {
		s := gocron.NewScheduler()
		s.Every(1).Hour().From(gocron.NextTick()).Do(AggloAndCommitMetrics)
		// s.Every(3).Minute().From(gocron.NextTick()).Do(AggloMetrics)

		s.Start()

		utils.Log("Metrics: Agglomeration Initialized")
	}()
}