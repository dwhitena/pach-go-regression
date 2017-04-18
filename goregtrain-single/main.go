// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

// Sample program to train a regression model.
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/sajari/regression"
)

// ModelInfo includes the information about the
// model that is output from the training.
type ModelInfo struct {
	Intercept    float64           `json:"intercept"`
	Coefficients []CoefficientInfo `json:"coefficients"`
}

// CoefficientInfo include information about a
// particular model coefficient.
type CoefficientInfo struct {
	Name        string  `json:"name"`
	Coefficient float64 `json:"coefficient"`
}

func main() {

	// Open the training dataset file.
	f, err := os.Open("../data/diabetes.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Create a new CSV reader reading from the opened file.
	reader := csv.NewReader(f)

	// Read in all of the CSV records
	reader.FieldsPerRecord = 11
	trainingData, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// In this case we are going to try and model our disease measure
	// y by the bmi feature plust an intercept.  As such, let's create
	// the struct needed to train a model using github.com/sajari/regression.
	var r regression.Regression
	r.SetObserved("diabetes progression")
	r.SetVar(0, "bmi")

	// Loop of records in the CSV, adding the training data to the regression value.
	for i, record := range trainingData {

		// Skip the header.
		if i == 0 {
			continue
		}

		// Parse the diabetes progression measure, or "y".
		yVal, err := strconv.ParseFloat(record[10], 64)
		if err != nil {
			log.Fatal(err)
		}

		// Parse the bmi value.
		bmiVal, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatal(err)
		}

		// Add these points to the regression value.
		r.Train(regression.DataPoint(yVal, []float64{bmiVal}))
	}

	// Train/fit the regression model.
	r.Run()

	// Output the trained model parameters to stdout.
	fmt.Printf("\nRegression Formula:\n%v\n\n", r.Formula)

	// Fill in the model information.
	modelInfo := ModelInfo{
		Intercept: r.Coeff(0),
		Coefficients: []CoefficientInfo{
			CoefficientInfo{
				Name:        "bmi",
				Coefficient: r.Coeff(1),
			},
		},
	}

	// Marshal the model information.
	outputData, err := json.MarshalIndent(modelInfo, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	// Save the marshalled output to a file.
	if err := ioutil.WriteFile("model.json", outputData, 0644); err != nil {
		log.Fatal(err)
	}
}
