package controllers

import (
	"encoding/json"

	"github.com/casosorg/casos/object"
)

// GetTrivyScanResults godoc
// @router /api/get-trivy-scan-results [get]
func (c *ApiController) GetTrivyScanResults() {
	results, err := object.GetTrivyScanResults()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(results)
}

// TriggerTrivyScan godoc
// @router /api/trigger-trivy-scan [post]
func (c *ApiController) TriggerTrivyScan() {
	var body struct {
		Image string `json:"image"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil || body.Image == "" {
		c.ResponseError("image is required")
		return
	}
	result, err := object.RunScanSync(body.Image)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}

// DeleteTrivyScanResult godoc
// @router /api/delete-trivy-scan-result [post]
func (c *ApiController) DeleteTrivyScanResult() {
	var body struct {
		Id int64 `json:"id"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &body); err != nil {
		c.ResponseError("invalid request body")
		return
	}
	if err := object.DeleteTrivyScanResult(body.Id); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
