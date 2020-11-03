package leancloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/levigross/grequests"
)

type ObjectRef struct {
	c     *Client
	class string
	ID    string
}

func (client *Client) Object(name, id string) *ObjectRef {
	return &ObjectRef{
		c:     client,
		class: name,
		ID:    id,
	}
}

func (ref *ObjectRef) Get(authOptions ...AuthOption) (*Object, error) {
	decodedObject, err := objectGet(ref, authOptions...)
	if err != nil {
		return nil, err
	}

	object, ok := decodedObject.(*Object)
	if !ok {
		return nil, fmt.Errorf("unexpected error when parse Object from response: want type *Object but %v", reflect.TypeOf(decodedObject))
	}
	return object, nil
}

func (ref *ObjectRef) Set(field string, value interface{}, authOptions ...AuthOption) error {
	if ref.ID == "" {
		return fmt.Errorf("no reference to object")
	}

	if err := objectSet(ref, field, value, authOptions...); err != nil {
		return err
	}

	return nil
}

func (ref *ObjectRef) Update(data map[string]interface{}, authOptions ...AuthOption) error {
	if ref.ID == "" {
		return fmt.Errorf("no reference to object")
	}

	if err := objectUpdate(ref, data, authOptions...); err != nil {
		return err
	}

	return nil
}

func (ref *ObjectRef) UpdateWithQuery(data map[string]interface{}, query *Query, authOptions ...AuthOption) error {
	// TODO
	return nil
}

func (ref *ObjectRef) Destroy(authOptions ...AuthOption) error {
	if ref.ID == "" {
		return fmt.Errorf("no reference to object")
	}

	if err := objectDestroy(ref); err != nil {
		return err
	}

	return nil
}

func objectCreate(class interface{}, data interface{}, authOptions ...AuthOption) (interface{}, error) {
	path := "/1.1/"
	var c *Client
	var options *grequests.RequestOptions

	switch v := class.(type) {
	case *Class:
		path = fmt.Sprint(path, "classes/", v.Name)
		c = v.c
		options = c.getRequestOptions()
		options.JSON = encodeObject(data)
		break
	case *Users:
		path = fmt.Sprint(path, "users")
		c = v.c
		options = c.getRequestOptions()
		options.JSON = data
		break
	}

	resp, err := c.request(ServiceAPI, MethodPost, path, options, authOptions...)
	if err != nil {
		return nil, err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return nil, err
	}
	switch v := class.(type) {
	case *Class:
		objectID, ok := respJSON["objectId"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected error when parse objectId from response: want type string but %v", reflect.TypeOf(respJSON["objectId"]))
		}
		return &ObjectRef{
			ID:    objectID,
			class: v.Name,
			c:     c,
		}, nil
	case *Users:
		return decodeUser(respJSON)
	}

	return nil, nil
}

func objectGet(ref interface{}, authOptions ...AuthOption) (interface{}, error) {
	path := "/1.1/"
	var c *Client

	switch v := ref.(type) {
	case *ObjectRef:
		path = fmt.Sprint(path, "classes/", v.class, "/", v.ID)
		c = v.c
		break
	case *UserRef:
		path = fmt.Sprint(path, "users/", v.ID)
		c = v.c
		break
	}

	resp, err := c.request(ServiceAPI, MethodGet, path, c.getRequestOptions(), authOptions...)
	if err != nil {
		return nil, err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return nil, err
	}

	switch ref.(type) {
	case *ObjectRef:
		return decodeObject(respJSON)
	case *UserRef:
		return decodeUser(respJSON)
	}

	return nil, nil
}

func objectSet(ref interface{}, field string, data interface{}, authOptions ...AuthOption) error {
	path := "/1.1/"
	var c *Client

	switch v := ref.(type) {
	case *ObjectRef:
		path = fmt.Sprint(path, "classes/", v.class, "/", v.ID)
		c = v.c
		break
	case *UserRef:
		path = fmt.Sprint(path, "users/", v.ID)
		c = v.c
		break
	}

	options := c.getRequestOptions()
	options.JSON = encode(map[string]interface{}{
		field: data,
	})

	resp, err := c.request(ServiceAPI, MethodPut, path, options, authOptions...)
	if err != nil {
		return err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return err
	}

	return nil
}

func objectUpdate(ref interface{}, data map[string]interface{}, authOptions ...AuthOption) error {
	path := "/1.1/"
	var c *Client

	switch v := ref.(type) {
	case *ObjectRef:
		path = fmt.Sprint(path, "classes/", v.class, "/", v.ID)
		c = v.c
		break
	case *UserRef:
		path = fmt.Sprint(path, "users/", v.ID)
		c = v.c
		break
	}

	options := c.getRequestOptions()
	options.JSON = encode(data)

	resp, err := c.request(ServiceAPI, MethodPut, path, options, authOptions...)
	if err != nil {
		return err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return err
	}

	return nil
}

func objectDestroy(ref interface{}, authOptions ...AuthOption) error {
	path := "/1.1/"
	var c *Client

	switch v := ref.(type) {
	case *ObjectRef:
		path = fmt.Sprint(path, "classes/", v.class, "/", v.ID)
		c = v.c
		break
	case *UserRef:
		path = fmt.Sprint(path, "users/", v.ID)
		c = v.c
		break
	}

	resp, err := c.request(ServiceAPI, MethodDelete, path, c.getRequestOptions(), authOptions...)
	if err != nil {
		return err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return err
	}

	return nil
}

func extracMetadata(respBody []byte) (id, token string, createdAt, updatedAt time.Time, fields map[string]interface{}, err error) {
	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(respBody, &respJSON); err != nil {
		return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse response body %w", err)
	}

	ok := false
	if respJSON["objectId"] != nil {
		id, ok = respJSON["objectId"].(string)
		if !ok {
			return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse objectId from response")
		}
	}

	if respJSON["sessionToken"] != nil {
		token, ok = respJSON["sessionToken"].(string)
		if !ok {
			return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse sessionToken from response")
		}
	}

	if respJSON["createdAt"] != nil {
		dateStr, ok := respJSON["createdAt"].(string)
		if !ok {
			return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse createdAt from response")
		}

		if dateStr != "" {
			date, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse createdAt from response")
			}
			createdAt = date
		}
	}

	if respJSON["updatedAt"] != nil {
		dateStr, ok := respJSON["updatedAt"].(string)
		if !ok {
			return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse updatedAt from response")
		}

		if dateStr != "" {
			date, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return "", "", time.Time{}, time.Time{}, nil, fmt.Errorf("unable to parse updatedAt from response")
			}
			updatedAt = date
		}
	}

	return id, token, createdAt, updatedAt, respJSON, nil
}
