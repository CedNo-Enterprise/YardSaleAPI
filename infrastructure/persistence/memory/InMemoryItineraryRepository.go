package memory

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/itinerary"
	"context"
)

type InMemoryItineraryRepository struct {
	itineraries []itinerary.Itinerary
}

// copyItinerary detaches an itinerary from the store so a caller mutating what it
// was handed cannot reach back into the slice the repository holds.
func copyItinerary(i itinerary.Itinerary) itinerary.Itinerary {
	stops := make([]itinerary.Stop, len(i.Stops()))
	copy(stops, i.Stops())

	return *itinerary.HydrateItinerary(
		i.Id(), i.UserId(), i.Name(), i.Description(), i.Date(),
		i.StartLatitude(), i.StartLongitude(), stops, i.CreatedAt(),
	)
}

func (repo *InMemoryItineraryRepository) indexOf(id string) int {
	for index, value := range repo.itineraries {
		if value.Id() == id {
			return index
		}
	}

	return -1
}

func (repo *InMemoryItineraryRepository) Create(ctx context.Context, i *itinerary.Itinerary) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if repo.indexOf(i.Id()) != -1 {
		return apperror.Conflict("itinerary already exists", nil)
	}

	repo.itineraries = append(repo.itineraries, copyItinerary(*i))
	return nil
}

func (repo *InMemoryItineraryRepository) GetById(ctx context.Context, id string) (*itinerary.Itinerary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	index := repo.indexOf(id)
	if index == -1 {
		return nil, apperror.NotFound("itinerary not found", nil)
	}

	found := copyItinerary(repo.itineraries[index])
	return &found, nil
}

func (repo *InMemoryItineraryRepository) GetByUserId(ctx context.Context, userId string) ([]itinerary.Itinerary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	found := make([]itinerary.Itinerary, 0, len(repo.itineraries))
	for _, value := range repo.itineraries {
		if value.UserId() == userId {
			found = append(found, copyItinerary(value))
		}
	}

	return found, nil
}

func (repo *InMemoryItineraryRepository) Update(ctx context.Context, i *itinerary.Itinerary) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(i.Id())
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	// The database repository writes scalar columns only, so preserve the stored
	// stops rather than taking whatever the caller's copy happens to hold.
	stored := repo.itineraries[index]
	updated := *itinerary.HydrateItinerary(
		i.Id(), i.UserId(), i.Name(), i.Description(), i.Date(),
		i.StartLatitude(), i.StartLongitude(), stored.Stops(), stored.CreatedAt(),
	)
	repo.itineraries[index] = updated

	return nil
}

func (repo *InMemoryItineraryRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(id)
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	repo.itineraries = append(repo.itineraries[:index], repo.itineraries[index+1:]...)
	return nil
}

func (repo *InMemoryItineraryRepository) AddStop(ctx context.Context, s *itinerary.Stop) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(s.ItineraryId())
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	if repo.itineraries[index].HasStop(s.SaleId()) {
		return apperror.Conflict("sale already in itinerary", nil)
	}

	repo.itineraries[index].AppendStop(s.SaleId())
	return nil
}

func (repo *InMemoryItineraryRepository) RemoveStop(ctx context.Context, itineraryId string, saleId string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(itineraryId)
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	if !repo.itineraries[index].RemoveStopBySaleId(saleId) {
		return apperror.NotFound("stop not found", nil)
	}

	return nil
}

func (repo *InMemoryItineraryRepository) ReorderStops(ctx context.Context, i *itinerary.Itinerary) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(i.Id())
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	repo.itineraries[index] = copyItinerary(*i)
	return nil
}

func (repo *InMemoryItineraryRepository) UpdateStopStatus(
	ctx context.Context, itineraryId string, saleId string, status itinerary.StopStatus,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	index := repo.indexOf(itineraryId)
	if index == -1 {
		return apperror.NotFound("itinerary not found", nil)
	}

	if !repo.itineraries[index].SetStopStatus(saleId, status) {
		return apperror.NotFound("stop not found", nil)
	}

	return nil
}
