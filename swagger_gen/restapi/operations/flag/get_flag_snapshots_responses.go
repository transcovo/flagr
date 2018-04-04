// Code generated by go-swagger; DO NOT EDIT.

package flag

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/checkr/flagr/swagger_gen/models"
)

// GetFlagSnapshotsOKCode is the HTTP code returned for type GetFlagSnapshotsOK
const GetFlagSnapshotsOKCode int = 200

/*GetFlagSnapshotsOK returns the flag snapshots

swagger:response getFlagSnapshotsOK
*/
type GetFlagSnapshotsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.FlagSnapshot `json:"body,omitempty"`
}

// NewGetFlagSnapshotsOK creates GetFlagSnapshotsOK with default headers values
func NewGetFlagSnapshotsOK() *GetFlagSnapshotsOK {

	return &GetFlagSnapshotsOK{}
}

// WithPayload adds the payload to the get flag snapshots o k response
func (o *GetFlagSnapshotsOK) WithPayload(payload []*models.FlagSnapshot) *GetFlagSnapshotsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get flag snapshots o k response
func (o *GetFlagSnapshotsOK) SetPayload(payload []*models.FlagSnapshot) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFlagSnapshotsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*models.FlagSnapshot, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

/*GetFlagSnapshotsDefault generic error response

swagger:response getFlagSnapshotsDefault
*/
type GetFlagSnapshotsDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetFlagSnapshotsDefault creates GetFlagSnapshotsDefault with default headers values
func NewGetFlagSnapshotsDefault(code int) *GetFlagSnapshotsDefault {
	if code <= 0 {
		code = 500
	}

	return &GetFlagSnapshotsDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get flag snapshots default response
func (o *GetFlagSnapshotsDefault) WithStatusCode(code int) *GetFlagSnapshotsDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get flag snapshots default response
func (o *GetFlagSnapshotsDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get flag snapshots default response
func (o *GetFlagSnapshotsDefault) WithPayload(payload *models.Error) *GetFlagSnapshotsDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get flag snapshots default response
func (o *GetFlagSnapshotsDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFlagSnapshotsDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
