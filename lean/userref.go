package lean

import "time"

type UserRef struct {
	c     *Client
	class string
	ID    string
	token string
}

type signupResponse struct {
	SessionToken string    `json:"sessionToken"`
	CreatedAt    time.Time `json:"createdAt"`
	ObjectID     string    `json:"objectId"`
}

type signinResponse struct {
	SessionToken        string    `json:"sessionToken"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Phone               string    `json:"phone"`
	ObjectID            string    `json:"objectId"`
	Username            string    `json:"username"`
	CreatedAt           time.Time `json:"createdAt"`
	EamilVerified       bool      `json:"emailVerified"`
	MobilePhoneVerified bool      `json:"mobilePhoneVerified"`
}

func (client *Client) User(id string) *UserRef {
	// TODO
	return nil
}

func (*UserRef) Login() error {
	// TODO
	return nil
}

func (r *UserRef) Signup() error {
	// TODO
	return nil
}

func (r *UserRef) ResetToken() error {
	// TODO
	return nil
}

func (r *UserRef) VerifyEmail() error {
	// TODO
	return nil
}

func (r *UserRef) VerifyPhoneNumber() error {
	// TODO
	return nil
}

func (r *UserRef) ResetPassword() error {
	// TODO
	return nil
}

func (r *UserRef) Get(authOption ...AuthOption) (*User, error) {
	// TODO
	return nil, nil
}

func (r *UserRef) Set(field string, value interface{}, authOption ...AuthOption) error {
	// TODO
	return nil
}

func (r *UserRef) Update(data map[string]interface{}, authOption ...AuthOption) error {
	// TODO
	return nil
}

func (r *UserRef) UpdateWithQuery(data map[string]interface{}, query *Query, authOption ...AuthOption) error {
	// TODO
	return nil
}

func (r *UserRef) Delete() error {
	// TODO
	return nil
}

func (r *UserRef) getSessionToken() string {
	return r.token
}