package services

import (
	"GarageSaleAPI/application/server/apperror"
	"GarageSaleAPI/domain/itinerary"
	"GarageSaleAPI/domain/sale"
	"GarageSaleAPI/interfaces/requests"
	"context"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ItineraryService struct {
	itineraryRepository itinerary.ItineraryRepository
	saleRepository      sale.SaleRepository
}

func NewItineraryService(
	itineraryRepository itinerary.ItineraryRepository, saleRepository sale.SaleRepository,
) *ItineraryService {
	return &ItineraryService{
		itineraryRepository: itineraryRepository,
		saleRepository:      saleRepository,
	}
}

// validateStartPair rejects a half-specified start point. Optimizing a route from
// one coordinate would silently produce a meaningless order.
func validateStartPair(latitude *float64, longitude *float64) error {
	if (latitude == nil) != (longitude == nil) {
		return apperror.Invalid("startLatitude and startLongitude must be provided together", nil)
	}

	return nil
}

func validateItinerary(itineraryDTO requests.ItineraryRequest) error {
	validate := validator.New()
	err := validate.Struct(itineraryDTO)
	if err != nil {
		return apperror.Invalid("invalid itinerary", err)
	}

	return validateStartPair(itineraryDTO.StartLatitude, itineraryDTO.StartLongitude)
}

func validateUpdateItinerary(updateDTO requests.UpdateItineraryRequest) error {
	validate := validator.New()
	err := validate.Struct(updateDTO)
	if err != nil {
		return apperror.Invalid("invalid itinerary", err)
	}

	// An entirely empty patch is a client mistake, not a silent no-op.
	if updateDTO.Name == nil && updateDTO.Description == nil && updateDTO.Date == nil &&
		updateDTO.StartLatitude == nil && updateDTO.StartLongitude == nil {
		return apperror.Invalid("no fields to update", nil)
	}

	return validateStartPair(updateDTO.StartLatitude, updateDTO.StartLongitude)
}

func validateAddStop(stopDTO requests.AddStopRequest) error {
	validate := validator.New()
	err := validate.Struct(stopDTO)
	if err != nil {
		return apperror.Invalid("invalid stop", err)
	}

	return nil
}

func validateReorderStops(reorderDTO requests.ReorderStopsRequest) error {
	validate := validator.New()
	err := validate.Struct(reorderDTO)
	if err != nil {
		return apperror.Invalid("invalid reorder", err)
	}

	return validateStartPair(reorderDTO.StartLatitude, reorderDTO.StartLongitude)
}

func validateStopStatus(statusDTO requests.UpdateStopStatusRequest) error {
	validate := validator.New()
	err := validate.Struct(statusDTO)
	if err != nil {
		return apperror.Invalid("invalid stop status", err)
	}

	return nil
}

// requireOwnedItinerary loads an itinerary and asserts the caller owns it. Every
// mutating operation goes through it so ownership cannot be skipped by adding a
// new endpoint later.
func (service *ItineraryService) requireOwnedItinerary(
	ctx context.Context, itineraryId string, userId string,
) (*itinerary.Itinerary, error) {
	if err := requireUuid(itineraryId, "itinerary not found"); err != nil {
		return nil, err
	}

	i, err := service.itineraryRepository.GetById(ctx, itineraryId)
	if err != nil {
		return nil, err
	}

	// Reads are public, so the itinerary's existence is not a secret and a
	// truthful 403 leaks nothing a masking 404 would hide.
	if i.UserId() != userId {
		err = apperror.Forbidden("itinerary does not belong to user", nil)
		slog.Error("ownership check failed", "itineraryId", itineraryId, "userId", userId)
		return nil, err
	}

	return i, nil
}

// coordinatesFor maps each sale that exists to its location. Sales missing from
// the result are treated as unlocatable by the optimizer rather than as an error,
// so a route survives a sale disappearing underneath it.
func (service *ItineraryService) coordinatesFor(
	ctx context.Context, saleIds []string,
) (map[string]itinerary.GeoPoint, error) {
	if len(saleIds) == 0 {
		return map[string]itinerary.GeoPoint{}, nil
	}

	sales, err := service.saleRepository.GetByIds(ctx, saleIds)
	if err != nil {
		return nil, err
	}

	coordinates := make(map[string]itinerary.GeoPoint, len(sales))
	for _, s := range sales {
		saleAddress := s.Address()
		coordinates[s.Id()] = itinerary.GeoPoint{
			Latitude:  saleAddress.Latitude(),
			Longitude: saleAddress.Longitude(),
		}
	}

	return coordinates, nil
}

// requireSalesExist rejects sale ids that do not resolve to a sale, naming the
// first one that did not so the caller can fix their request.
func requireSalesExist(saleIds []string, coordinates map[string]itinerary.GeoPoint) error {
	for _, saleId := range saleIds {
		if _, ok := coordinates[saleId]; !ok {
			return apperror.Invalid("unknown sale id: "+saleId, nil)
		}
	}

	return nil
}

func floatOrZero(value *float64) float64 {
	if value == nil {
		return 0
	}

	return *value
}

func (service *ItineraryService) AddItinerary(
	ctx context.Context, userId string, itineraryDTO requests.ItineraryRequest,
) (*string, error) {
	err := validateItinerary(itineraryDTO)
	if err != nil {
		slog.Error("error adding itinerary", "err", err.Error())
		return nil, err
	}

	coordinates, err := service.coordinatesFor(ctx, itineraryDTO.SaleIds)
	if err != nil {
		slog.Error("error adding itinerary", "err", err.Error())
		return nil, err
	}
	if err = requireSalesExist(itineraryDTO.SaleIds, coordinates); err != nil {
		slog.Error("error adding itinerary", "err", err.Error())
		return nil, err
	}

	start := itinerary.GeoPoint{
		Latitude:  floatOrZero(itineraryDTO.StartLatitude),
		Longitude: floatOrZero(itineraryDTO.StartLongitude),
	}

	itineraryId := uuid.NewString()
	i := itinerary.CreateItinerary(
		itineraryId, userId, itineraryDTO.Name, itineraryDTO.Description,
		itineraryDTO.Date, start.Latitude, start.Longitude, time.Now(),
	)
	for _, saleId := range itineraryDTO.SaleIds {
		i.AppendStop(saleId)
	}
	i.ReorderFrom(start, coordinates)

	err = service.itineraryRepository.Create(ctx, i)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &itineraryId, nil
}

func (service *ItineraryService) GetItineraryById(
	ctx context.Context, itineraryId string,
) (*itinerary.Itinerary, error) {
	if err := requireUuid(itineraryId, "itinerary not found"); err != nil {
		return nil, err
	}

	i, err := service.itineraryRepository.GetById(ctx, itineraryId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return i, nil
}

func (service *ItineraryService) GetItinerariesByUserId(
	ctx context.Context, userId string,
) ([]itinerary.Itinerary, error) {
	itineraries, err := service.itineraryRepository.GetByUserId(ctx, userId)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return itineraries, nil
}

func (service *ItineraryService) UpdateItinerary(
	ctx context.Context, itineraryId string, userId string,
	updateDTO requests.UpdateItineraryRequest,
) (*itinerary.Itinerary, error) {
	err := validateUpdateItinerary(updateDTO)
	if err != nil {
		slog.Error("error updating itinerary", "err", err.Error())
		return nil, err
	}

	i, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return nil, err
	}

	if updateDTO.Name != nil {
		i.Rename(*updateDTO.Name)
	}
	if updateDTO.Description != nil {
		i.Describe(*updateDTO.Description)
	}
	if updateDTO.Date != nil {
		i.Reschedule(*updateDTO.Date)
	}
	// Moving the start point does not re-sequence the route; reordering is an
	// explicit action so a PATCH never surprises the caller with new positions.
	if updateDTO.StartLatitude != nil && updateDTO.StartLongitude != nil {
		i.SetStart(*updateDTO.StartLatitude, *updateDTO.StartLongitude)
	}

	err = service.itineraryRepository.Update(ctx, i)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return i, nil
}

func (service *ItineraryService) DeleteItinerary(
	ctx context.Context, itineraryId string, userId string,
) error {
	_, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return err
	}

	err = service.itineraryRepository.Delete(ctx, itineraryId)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *ItineraryService) AddStop(
	ctx context.Context, itineraryId string, userId string, stopDTO requests.AddStopRequest,
) (*itinerary.Itinerary, error) {
	err := validateAddStop(stopDTO)
	if err != nil {
		slog.Error("error adding stop", "err", err.Error())
		return nil, err
	}

	i, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return nil, err
	}

	if i.HasStop(stopDTO.SaleId) {
		err = apperror.Conflict("sale already in itinerary", nil)
		slog.Error("error adding stop", "err", err.Error())
		return nil, err
	}

	// Confirms the sale is real before it joins the route.
	coordinates, err := service.coordinatesFor(ctx, []string{stopDTO.SaleId})
	if err != nil {
		slog.Error("error adding stop", "err", err.Error())
		return nil, err
	}
	if err = requireSalesExist([]string{stopDTO.SaleId}, coordinates); err != nil {
		slog.Error("error adding stop", "err", err.Error())
		return nil, err
	}

	// Appended rather than re-optimized so adding a sale stays one insert. The
	// caller reorders when it wants a fresh route.
	stop := i.AppendStop(stopDTO.SaleId)

	err = service.itineraryRepository.AddStop(ctx, &stop)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	// The repository assigns the position under a lock, so report what was
	// stored rather than the one guessed from a read that may now be stale.
	return service.itineraryRepository.GetById(ctx, itineraryId)
}

func (service *ItineraryService) RemoveStop(
	ctx context.Context, itineraryId string, userId string, saleId string,
) error {
	i, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return err
	}

	if !i.HasStop(saleId) {
		err = apperror.NotFound("stop not found", nil)
		slog.Error("error removing stop", "err", err.Error())
		return err
	}

	err = service.itineraryRepository.RemoveStop(ctx, itineraryId, saleId)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (service *ItineraryService) ReorderStops(
	ctx context.Context, itineraryId string, userId string,
	reorderDTO requests.ReorderStopsRequest,
) (*itinerary.Itinerary, error) {
	err := validateReorderStops(reorderDTO)
	if err != nil {
		slog.Error("error reordering stops", "err", err.Error())
		return nil, err
	}

	i, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return nil, err
	}

	// Omitted coordinates reuse the start already stored on the itinerary.
	start := i.Start()
	if reorderDTO.StartLatitude != nil && reorderDTO.StartLongitude != nil {
		start = itinerary.GeoPoint{
			Latitude:  *reorderDTO.StartLatitude,
			Longitude: *reorderDTO.StartLongitude,
		}
	}

	saleIds := make([]string, 0, len(i.Stops()))
	for _, stop := range i.Stops() {
		saleIds = append(saleIds, stop.SaleId())
	}

	coordinates, err := service.coordinatesFor(ctx, saleIds)
	if err != nil {
		slog.Error("error reordering stops", "err", err.Error())
		return nil, err
	}

	i.ReorderFrom(start, coordinates)

	err = service.itineraryRepository.ReorderStops(ctx, i)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return i, nil
}

func (service *ItineraryService) SetStopStatus(
	ctx context.Context, itineraryId string, userId string, saleId string,
	statusDTO requests.UpdateStopStatusRequest,
) (*itinerary.Itinerary, error) {
	err := validateStopStatus(statusDTO)
	if err != nil {
		slog.Error("error setting stop status", "err", err.Error())
		return nil, err
	}

	i, err := service.requireOwnedItinerary(ctx, itineraryId, userId)
	if err != nil {
		return nil, err
	}

	status := itinerary.StopStatus(statusDTO.Status)
	if !i.SetStopStatus(saleId, status) {
		err = apperror.NotFound("stop not found", nil)
		slog.Error("error setting stop status", "err", err.Error())
		return nil, err
	}

	err = service.itineraryRepository.UpdateStopStatus(ctx, itineraryId, saleId, status)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return i, nil
}
