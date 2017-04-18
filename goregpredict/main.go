// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

// Sample program to make a prediction based on a persisted regression model.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	inModelDir = "../goregtrain-multi"
	inVarDir   = "../data/test"
	outDir     = ""
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

// PredictionData includes the data necessary to make
// a prediction and encodes the output prediction.
type PredictionData struct {
	Prediction      float64          `json:"predicted_diabetes_progression"`
	IndependentVars []IndependentVar `json:"independent_variables"`
}

// IndependentVar include information about and a
// value for an independent variable.
type IndependentVar struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func main() {

	// Load the model file.
	f, err := ioutil.ReadFile(filepath.Join(inModelDir, "model.json"))
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the model information.
	var modelInfo ModelInfo
	if err := json.Unmarshal(f, &modelInfo); err != nil {
		log.Fatal(err)
	}

	// Walk over files in the input.
	if err := filepath.Walk(inVarDir, func(path string, info os.FileInfo, err error) error {

		// Skip any directories.
		if info.IsDir() {
			return nil
		}

		// Open any files.
		f, err := ioutil.ReadFile(filepath.Join(inVarDir, info.Name()))
		if err != nil {
			return err
		}

		// Unmarshal the independent variables.
		var predictionData PredictionData
		if err := json.Unmarshal(f, &predictionData); err != nil {
			return err
		}

		// Make the prediction.
		if err := Predict(&modelInfo, &predictionData); err != nil {
			return err
		}

		// Marshal the prediction data.
		outputData, err := json.MarshalIndent(predictionData, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		// Save the marshalled output to a file.
		if err := ioutil.WriteFile(filepath.Join(outDir, info.Name()), outputData, 0644); err != nil {
			log.Fatal(err)
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}

// Predict makes a prediction based on input JSON.
func Predict(modelInfo *ModelInfo, predictionData *PredictionData) error {

	// Initialize the prediction value
	// to the intercept.
	prediction := modelInfo.Intercept

	// Create a map of independent variable coefficients.
	coeffs := make(map[string]float64)
	varNames := make([]string, len(modelInfo.Coefficients))
	for idx, coeff := range modelInfo.Coefficients {
		coeffs[coeff.Name] = coeff.Coefficient
		varNames[idx] = coeff.Name
	}

	// Create a map of the independent variable values.
	varVals := make(map[string]float64)
	for _, indVar := range predictionData.IndependentVars {
		varVals[indVar.Name] = indVar.Value
	}

	// Loop over the independent variables.
	for _, varName := range varNames {

		// Get the coefficient.
		coeff, ok := coeffs[varName]
		if !ok {
			return fmt.Errorf("Could not find model coefficient %s", varName)
		}

		// Get the variable value.
		val, ok := varVals[varName]
		if !ok {
			return fmt.Errorf("Expected a value for variable %s", varName)
		}

		// Add to the prediction.
		prediction = prediction + coeff*val
	}

	// Add the prediction to the prediction data.
	predictionData.Prediction = prediction

	return nil
}
