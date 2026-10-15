package schema

// ProfileInput distinguishes omitted values from explicit null values on PATCH.
type ProfileInput struct {
	DisplayName OptionalString `json:"displayName"`
	Username    OptionalString `json:"username"`
	Bio         OptionalString `json:"bio"`
	AvatarURL   OptionalString `json:"avatarUrl"`
}

type UserStatusInput struct {
	Status string `json:"status"`
}
