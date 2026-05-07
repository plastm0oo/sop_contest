package service

type HealthResponse struct {
	Status string `json:"status"`
}

type TeacherListParams struct {
	Q       string
	Faculty string
	Limit   int
	Offset  int
}

type TeacherListItem struct {
	ID           int64   `db:"id" json:"id"`
	FullName     string  `db:"full_name" json:"full_name"`
	Faculty      string  `db:"faculty" json:"faculty"`
	ReviewsCount int64   `db:"reviews_count" json:"reviews_count"`
	AvgRating    float64 `db:"avg_rating" json:"avg_rating"`
}

type TeacherListResponse struct {
	Items  []TeacherListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
