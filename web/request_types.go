package web

type CreateSegmentBody struct {
	// Name of a new segment. You cannot use a name of a currently or
	// previously existing segment. The name cannot be an empty string.
	Name string `json:"name"`

	// Probability that new known users will also be part of this segment
	// without explicitly requesting so. Default: 0.
	//
	// _percent_% previously known users will get assigned to this segment
	// upon this request. All new known users will be assigned to this segment
	// with _percent_% probability.
	Percent int32 `json:"percent,omitempty"`
}

type DeleteSegmentBody struct {
	// Name of the segment to delete.
	Name string `json:"name"`
}

type GetSegmentsBody struct {
	Id int32 `json:"id"`
}

type HistoryBody struct {
	Year int32 `json:"year"`

	Month int32 `json:"month"`
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
