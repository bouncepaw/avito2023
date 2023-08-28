package web

type ResponseUsual struct {
	// Status of the operation. If `ok`, then the operation went correctly,
	// and you can ignore the `error` field. If `error`, an error occurred
	// which is specified in the `error` field.
	Status string `json:"status"`

	// Explanation of the error. Set only if `status` is `error`.
	//
	// Values:
	// * `name empty` means the passed name is an empty string.
	// * `name taken` means the provided name is taken already and cannot be used for new segments.
	// * `name free` means that no segment with the given name exists.
	// * `segment deleted` means that the segment is segment deleted.
	// * `bad percent` means the passed percent value is outside 0..100 range.
	//* Other values are internal or parsing errors.
	Error_ string `json:"error,omitempty"`
}

type ResponseGetSegments struct {
	Status string `json:"status"`

	Error_ string `json:"error,omitempty"`

	Segments []string `json:"segments,omitempty"`
}

type ResponseHistory struct {
	Status string `json:"status"`

	// Set if `status` is `error`. Possible values:
	// * `bad time` means the year or month you passed is invalid in general.
	// * Other values are internal or parsing errors.
	Error_ string `json:"error,omitempty"`

	// If `status` is `ok`, link starts with /. Request the file at the same
	// server. If `error`, this string is empty.
	Link string `json:"link,omitempty"`
}

type UpdateUserBody struct {
	Id int32 `json:"id"`
	// Segments to add the user to. Duplicates are ignored. For any given segment, if the user is already part of it, nothing happens and no error is returned.
	AddToSegments []string `json:"add_to_segments,omitempty"`
	// Segments to remove the user from. Duplicates are ignored. For any given segment, if the user is not part of it, nothing happend and no error is returned.
	RemoveFromSegments []string `json:"remove_from_segments,omitempty"`
	// Time to live. Seconds to wait before removing the user from all the `add_to_segments` segments.
	Ttl int32 `json:"ttl,omitempty"`
}
