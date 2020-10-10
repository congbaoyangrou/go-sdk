package leancloud

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/levigross/grequests"
)

type Query struct {
	c     *Client
	class *Class
	where map[string]interface{}
	order []string
	limit int
	skip  int
}

func (q *Query) Find(authOptions ...AuthOption) ([]Object, error) {
	respObjects, err := objectQuery(q, false, false, authOptions...)
	if err != nil {
		return nil, err
	}

	objects, ok := respObjects.([]Object)
	if !ok {
		return nil, fmt.Errorf("unable to parse objects from response")
	}

	return objects, nil
}

func (q *Query) First(authOptions ...AuthOption) (*Object, error) {
	respObjects, err := objectQuery(q, false, true, authOptions...)
	if err != nil {
		return nil, err
	}

	objects, ok := respObjects.([]Object)
	if !ok {
		return nil, fmt.Errorf("unable to parse object from response")
	}

	return &objects[0], nil
}

func (q *Query) Count(authOptions ...AuthOption) (int, error) {
	resp, err := objectQuery(q, true, false, authOptions...)
	if err != nil {
		return 0, err
	}

	count, ok := resp.(float64)
	if !ok {
		return 0, fmt.Errorf("unable to parse count from response")
	}

	return int(count), nil
}

func (q *Query) Skip(count int) *Query {
	q.skip = count
	return q
}

func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

func (q *Query) Order(keys ...string) *Query {
	q.order = keys
	return q
}

func (q *Query) EqualTo(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("", value, "")
	return q
}

func (q *Query) NotEqualTo(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("$ne", value, "")
	return q
}

func (q *Query) GreaterThan(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("$gt", value, "")
	return q
}

func (q *Query) GreaterThanOrEqualTo(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("$gte", value, "")
	return q
}

func (q *Query) LessThan(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("$lt", value, "")
	return q
}

func (q *Query) LessThanOrEqualTo(key string, value interface{}) *Query {
	q.where[key] = wrapCondition("$lte", value, "")
	return q
}

func (q *Query) In(key string, data interface{}) *Query {
	q.where[key] = wrapCondition("$in", data, "")
	return q
}

func (q *Query) NotIn(key string, data interface{}) *Query {
	q.where[key] = wrapCondition("$nin", data, "")
	return q
}

func (q *Query) Regexp(key, expr, options string) *Query {
	q.where[key] = wrapCondition("$regex", expr, options)
	return q
}

func (q *Query) Contains(key, substring string) *Query {
	q.Regexp(key, substring, "")
	return q
}

func (q *Query) ContainsAll(key string, objects interface{}) *Query {
	q.where[key] = wrapCondition("$all", objects, "")
	return q
}

func (q *Query) StartsWith(key, prefix string) *Query {
	q.Regexp(key, fmt.Sprint("^", prefix), "")
	return q
}

func wrapCondition(verb string, value interface{}, options string) interface{} {
	switch verb {
	case "$ne":
		fallthrough
	case "$lt":
		fallthrough
	case "$lte":
		fallthrough
	case "$gt":
		fallthrough
	case "$gte":
		fallthrough
	case "$in":
		fallthrough
	case "$nin":
		fallthrough
	case "$all":
		switch v := value.(type) {
		case time.Time:
			return map[string]interface{}{
				verb: encodeDate(v),
			}
		default:
			return map[string]interface{}{
				verb: value,
			}
		}
	case "$regex":
		return map[string]interface{}{
			"$regex":   value,
			"$options": options,
		}
	default:
		switch v := value.(type) {
		case time.Time:
			return encodeDate(v)
		default:
			return value
		}
	}
}

func objectQuery(query interface{}, count bool, first bool, authOptions ...AuthOption) (interface{}, error) {
	path := fmt.Sprint("/1.1/")
	var client *Client
	var options *grequests.RequestOptions
	params, err := wrapParams(query, count, first)
	if err != nil {
		return nil, err
	}

	switch v := query.(type) {
	case *Query:
		path = fmt.Sprint(path, "classes/", v.class.Name)
		options = v.c.GetRequestOptions()
		client = v.c
		break
	case *UserQuery:
		path = fmt.Sprint(path, "users/")
		options = v.c.GetRequestOptions()
		client = v.c
		break
	}

	options.Params = params

	resp, err := client.Request(ServiceAPI, MethodGet, path, options, authOptions...)
	if err != nil {
		return nil, err
	}

	respJSON := make(map[string]interface{})
	if err := json.Unmarshal(resp.Bytes(), &respJSON); err != nil {
		return nil, fmt.Errorf("unable to parse response %w", err)
	}

	if count {
		return respJSON["count"], nil
	}

	results := respJSON["results"].([]interface{})
	switch query.(type) {
	case *Query:
		var objects []Object
		for i := 0; i < len(results); i++ {
			result := results[i].(map[string]interface{})
			object := new(Object)
			DecodeObject(result, object)
			objects = append(objects, *object)
		}
		return objects, nil
	case *UserQuery:
		var users []User
		for i := 0; i < len(results); i++ {
			result := results[i].(map[string]interface{})
			user := new(User)
			DecodeObject(result, user)
			users = append(users, *user)
		}
		return users, nil
	}

	return nil, nil
}

func wrapParams(query interface{}, count, first bool) (map[string]string, error) {
	var where map[string]interface{}
	var order []string
	var skip, limit int

	switch v := query.(type) {
	case *Query:
		where = v.where
		order = v.order
		skip, limit = v.skip, v.limit
		break
	case *UserQuery:
		where = v.where
		order = v.order
		skip, limit = v.skip, v.limit
		break
	}

	whereString, err := json.Marshal(where)
	if err != nil {
		return nil, fmt.Errorf("unable to wrap params %w", err)
	}

	params := map[string]string{
		"where": string(whereString),
	}

	if skip != 0 {
		params["skip"] = fmt.Sprintf("%d", skip)
	}

	if limit != 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	if len(order) != 0 {
		params["order"] = strings.Join(order, ",")
	}

	if count {
		params["count"] = "1"
	}

	if first {
		params["limit"] = "1"
	}

	return params, nil
}
