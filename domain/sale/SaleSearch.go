package sale

import (
	"sort"
	"time"
)

const (
	DefaultSearchLimit = 20
	MaxSearchLimit     = 100
)

type SortOrder string

const (
	SortDate        SortOrder = "date"
	SortDateReverse SortOrder = "-date"
	SortCreated     SortOrder = "created"
)

// SearchCriteria describes a browse query. A nil date bound is open-ended and an
// empty status list means the default set rather than "match nothing", so the
// zero value is a usable first page.
type SearchCriteria struct {
	Statuses []Status
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     SortOrder
	Limit    int
	Offset   int
}

// DefaultSearchStatuses leaves out cancelled sales. Browse is a discovery
// surface and a cancelled sale is one that no longer happens; a caller that
// wants them asks for them by name.
func DefaultSearchStatuses() []Status {
	return []Status{StatusScheduled, StatusActive, StatusCompleted}
}

// Normalize fills in the defaults and bounds so the repositories and the service
// all work from the same criteria rather than each applying its own.
func (c SearchCriteria) Normalize() SearchCriteria {
	if len(c.Statuses) == 0 {
		c.Statuses = DefaultSearchStatuses()
	}

	switch c.Sort {
	case SortDate, SortDateReverse, SortCreated:
	default:
		c.Sort = SortDate
	}

	if c.Limit < 1 {
		c.Limit = DefaultSearchLimit
	}
	if c.Limit > MaxSearchLimit {
		c.Limit = MaxSearchLimit
	}
	if c.Offset < 0 {
		c.Offset = 0
	}

	return c
}

// Matches reports whether a sale satisfies every filter. The database repository
// spells the same predicate as SQL, so this is the definition both follow.
func (c SearchCriteria) Matches(s Sale) bool {
	if !containsStatus(c.Statuses, s.Status()) {
		return false
	}
	if c.DateFrom != nil && s.Date().Before(*c.DateFrom) {
		return false
	}
	if c.DateTo != nil && s.Date().After(*c.DateTo) {
		return false
	}

	return true
}

// Apply filters, orders and pages sales in memory. It exists so the in-memory
// repository can mirror the database one without restating the rules.
func Apply(sales []Sale, c SearchCriteria) []Sale {
	c = c.Normalize()

	matched := make([]Sale, 0, len(sales))
	for _, s := range sales {
		if c.Matches(s) {
			matched = append(matched, s)
		}
	}

	sortSales(matched, c.Sort)

	if c.Offset >= len(matched) {
		return []Sale{}
	}

	end := min(c.Offset+c.Limit, len(matched))

	return matched[c.Offset:end]
}

func containsStatus(statuses []Status, status Status) bool {
	for _, candidate := range statuses {
		if candidate == status {
			return true
		}
	}

	return false
}

// sortSales breaks ties on id so a page boundary cannot land differently between
// two requests that asked for the same thing.
func sortSales(sales []Sale, order SortOrder) {
	sort.SliceStable(sales, func(i, j int) bool {
		first, second := sales[i], sales[j]

		switch order {
		case SortDateReverse:
			if !first.Date().Equal(second.Date()) {
				return first.Date().After(second.Date())
			}
		case SortCreated:
			if !first.CreatedAt().Equal(second.CreatedAt()) {
				return first.CreatedAt().After(second.CreatedAt())
			}
		default:
			if !first.Date().Equal(second.Date()) {
				return first.Date().Before(second.Date())
			}
		}

		return first.Id() < second.Id()
	})
}
