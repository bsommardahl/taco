package handlers

import (
	"encoding/json"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sul-dlss-labs/taco/datautils"
	"github.com/sul-dlss-labs/taco/db"
	"github.com/sul-dlss-labs/taco/generated/models"
	"github.com/sul-dlss-labs/taco/generated/restapi/operations"
	"github.com/sul-dlss-labs/taco/identifier"
	"github.com/sul-dlss-labs/taco/validators"
)

// NewUpdateResource -- Accepts requests to update a resource.
func NewUpdateResource(database db.Database, validator validators.ResourceValidator) operations.UpdateResourceHandler {
	return &updateResourceEntry{database: database, validator: validator}
}

type updateResourceEntry struct {
	database  db.Database
	validator validators.ResourceValidator
}

// Handle the update resource request
func (d *updateResourceEntry) Handle(params operations.UpdateResourceParams) middleware.Responder {
	id := params.ID
	newResource := datautils.NewResource(params.Payload.(map[string]interface{}))

	if errors := d.validator.ValidateResource(newResource); errors != nil {
		return operations.NewUpdateResourceUnprocessableEntity().
			WithPayload(&models.ErrorResponse{Errors: *errors})
	}

	existingResource, err := d.database.RetrieveLatest(id)
	if err != nil {
		if _, ok := err.(*db.RecordNotFound); ok {
			return operations.NewUpdateResourceNotFound()
		}
		panic(err)
	}

	errors := d.verifyPayload(newResource, existingResource)
	if errors != nil {
		return operations.NewUpdateResourceUnprocessableEntity().WithPayload(&models.ErrorResponse{Errors: *errors})
	}

	v, _ := newResource.JSON["version"].(json.Number).Float64()
	version := int(v)
	if version > existingResource.Version() {
		d.buildNewResourceVersion(newResource, version, existingResource)
		response := datautils.JSONObject{"id": id}
		return operations.NewUpdateResourceOK().WithPayload(response)
	}

	// We need to ensure in this case that ID and externalID are NOT overwritten by the incoming JSON
	newResource = datautils.NewResource(d.mergeJSON(&existingResource.JSON, &newResource.JSON)).
		WithVersion(existingResource.Version()).
		WithID(existingResource.ID()).
		WithExternalIdentifier(existingResource.ExternalIdentifier())

	err = d.database.Insert(newResource)
	if err != nil {
		panic(err)
	}

	response := datautils.JSONObject{"id": id}
	return operations.NewUpdateResourceOK().WithPayload(response)
}

// Merges multiple JSONObjects. Overritting in order.
func (d *updateResourceEntry) mergeJSON(maps ...*datautils.JSONObject) datautils.JSONObject {
	result := make(datautils.JSONObject)
	for _, m := range maps {
		for k, v := range *m {
			switch v.(type) {
			case datautils.JSONObject:
				if _, ok := result[k]; ok {
					x := v.(datautils.JSONObject)
					result[k] = d.mergeJSON(result.GetObj(k), &x)
				} else {
					result[k] = v
				}
			default:
				result[k] = v
			}
		}
	}
	return result
}

func (d *updateResourceEntry) buildNewResourceVersion(newResource *datautils.Resource, version int, existingResource *datautils.Resource) {
	tacoIdentifier, err := identifier.NewUUIDService().Mint(newResource)
	if err != nil {
		panic(err)
	}

	newResource = datautils.NewResource(d.mergeJSON(&existingResource.JSON, &newResource.JSON)).
		WithID(tacoIdentifier).
		WithExternalIdentifier(existingResource.ExternalIdentifier()).
		WithPrecedingVersion(existingResource.ID()).
		WithVersion(version)

	err = d.database.Insert(newResource)
	if err != nil {
		panic(err)
	}

	deprecatedResource := datautils.NewResource(existingResource.JSON).
		WithFollowingVersion(tacoIdentifier)

	err = d.database.Insert(deprecatedResource)
	if err != nil {
		panic(err)
	}
}

func (d *updateResourceEntry) verifyPayload(newResource *datautils.Resource, existingResource *datautils.Resource) *models.ErrorResponseErrors {
	errors := models.ErrorResponseErrors{}
	if newResource.ExternalIdentifier() != existingResource.ExternalIdentifier() {
		errors = append(errors, &models.Error{Title: "Invalid Update Payload", Detail: "externalIdentifier in payload: does not match existing resource"})
		return &errors
	}
	if newResource.ID() != existingResource.ID() {
		errors = append(errors, &models.Error{Title: "Invalid Update Payload", Detail: "tacoIdentifier in payload: does not match existing resource"})
		return &errors
	}
	return nil
}
