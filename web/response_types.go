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
	Err string `json:"error,omitempty"`
}

type ResponseGetSegments struct {
	Status string `json:"status"`

	Err string `json:"error,omitempty"`

	Segments []string `json:"segments,omitempty"`
}

type ResponseHistory struct {
	Status string `json:"status"`

	// Set if `status` is `error`. Possible values:
	// * `bad time` means the year or month you passed is invalid in general.
	// * Other values are internal or parsing errors.
	Err string `json:"error,omitempty"`

	// If `status` is `ok`, link starts with /. Request the file at the same
	// server. If `error`, this string is empty.
	Link string `json:"link,omitempty"`
}
