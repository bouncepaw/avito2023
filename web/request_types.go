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
