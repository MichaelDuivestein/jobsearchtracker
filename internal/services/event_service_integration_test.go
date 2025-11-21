package services_test

import (
	"errors"
	"jobsearchtracker/internal/api/v1/requests"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupEventService(t *testing.T) (
	*services.EventService,
	*repositories.ApplicationRepository,
	*repositories.CompanyRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.ApplicationEventRepository,
	*repositories.CompanyEventRepository,
	*repositories.EventPersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventServiceTestContainer(t, *config)

	var eventService *services.EventService
	err := container.Invoke(func(service *services.EventService) {
		eventService = service
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var applicationEventRepository *repositories.ApplicationEventRepository
	err = container.Invoke(func(repository *repositories.ApplicationEventRepository) {
		applicationEventRepository = repository
	})
	assert.NoError(t, err)

	var companyEventRepository *repositories.CompanyEventRepository
	err = container.Invoke(func(repository *repositories.CompanyEventRepository) {
		companyEventRepository = repository
	})
	assert.NoError(t, err)

	var eventPersonRepository *repositories.EventPersonRepository
	err = container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
	})
	assert.NoError(t, err)

	return eventService,
		applicationRepository,
		companyRepository,
		eventRepository,
		personRepository,
		applicationEventRepository,
		companyEventRepository,
		eventPersonRepository
}

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldWork(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	insertedEvent, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, insertedEvent)

	assert.Equal(t, *createEvent.ID, insertedEvent.ID)
	assert.Equal(t, createEvent.EventType.String(), insertedEvent.EventType.String())
	assert.Equal(t, createEvent.Description, insertedEvent.Description)
	assert.Equal(t, createEvent.Notes, insertedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, insertedEvent.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, insertedEvent.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.UpdatedDate, insertedEvent.UpdatedDate)
}

func TestCreate_ShouldInsertEventWithOnlyRequiredFields(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	createdDateApproximation := time.Now()

	insertedEvent, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, insertedEvent)

	assert.NotNil(t, insertedEvent.ID)
	assert.Equal(t, createEvent.EventType.String(), insertedEvent.EventType.String())
	assert.Nil(t, insertedEvent.Description)
	assert.Nil(t, insertedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, insertedEvent.EventDate)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, insertedEvent.CreatedDate, time.Second)
	assert.Nil(t, insertedEvent.UpdatedDate)
}

// -------- GetEventByID tests: --------

func TestGetEventByID_ShouldWork(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 7, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 6, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, *createEvent.ID, event.ID)
	assert.Equal(t, createEvent.EventType.String(), event.EventType.String())
	assert.Equal(t, createEvent.Description, event.Description)
	assert.Equal(t, createEvent.Notes, event.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, event.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, event.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.UpdatedDate, event.UpdatedDate)
}

func TestGetEventByID_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	id := uuid.New()
	nilEvent, err := eventService.GetEventByID(&id)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		notFoundError.Error())
}

// -------- GetAllEvents - base tests: --------

func TestGetAllEvents_ShouldReturnAllEvents(t *testing.T) {
	eventService, _, _, eventRepository, _, _, _, _ := setupEventService(t)

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent1)
	assert.NoError(t, err)

	createEvent2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Equal(t, 2, len(events))

	assert.Equal(t, *createEvent1.ID, events[0].ID)
	assert.Equal(t, createEvent1.EventType.String(), events[0].EventType.String())
	assert.Equal(t, createEvent1.Description, events[0].Description)
	assert.Equal(t, createEvent1.Notes, events[0].Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent1.EventDate, events[0].EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.CreatedDate, events[0].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.UpdatedDate, events[0].UpdatedDate)

	assert.Equal(t, createEvent2.ID, events[1].ID)
	assert.Equal(t, createEvent2.EventType.String(), events[1].EventType.String())
	assert.Nil(t, createEvent2.Description)
	assert.Nil(t, createEvent2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createEvent2.EventDate, events[1].EventDate)
	assert.NotNil(t, createEvent2.CreatedDate)
	assert.Nil(t, createEvent2.UpdatedDate)
}

func TestGetAllEvents_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- GetAllEvents - Application tests: --------

func TestEventRepositoryGetAllEvents_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	eventService,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventService(t)

	// create events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// add two companies

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company1ID,
		RecruiterID:          &company2ID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company1ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate events and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, *createApplication1.ID, event1ID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, event1ID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, event2ID, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 2)

	assert.Equal(t, event1ID, events[0].ID)
	assert.Len(t, *(events[0]).Applications, 2)

	assert.Equal(t, application2ID, (*(*events[0]).Applications)[0].ID)

	event1Application2 := (*(*events[0]).Applications)[1]
	assert.Equal(t, *createApplication1.ID, event1Application2.ID)
	assert.Equal(t, createApplication1.CompanyID, event1Application2.CompanyID)
	assert.Equal(t, createApplication1.RecruiterID, event1Application2.RecruiterID)
	assert.Equal(t, createApplication1.JobTitle, event1Application2.JobTitle)
	assert.Equal(t, createApplication1.JobAdURL, event1Application2.JobAdURL)
	assert.Equal(t, createApplication1.Country, event1Application2.Country)
	assert.Equal(t, createApplication1.Area, event1Application2.Area)
	assert.Equal(t, createApplication1.RemoteStatusType.String(), event1Application2.RemoteStatusType.String())
	assert.Equal(t, createApplication1.WeekdaysInOffice, event1Application2.WeekdaysInOffice)
	assert.Equal(t, createApplication1.EstimatedCycleTime, event1Application2.EstimatedCycleTime)
	assert.Equal(t, createApplication1.EstimatedCommuteTime, event1Application2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.ApplicationDate, event1Application2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.CreatedDate, event1Application2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.UpdatedDate, event1Application2.UpdatedDate)

	assert.Len(t, *(events[1]).Applications, 1)
	assert.Equal(t, application2ID, (*(*events[1]).Applications)[0].ID)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplications(t *testing.T) {
	eventService, applicationRepository, companyRepository, eventRepository, _, _, _, _ := setupEventService(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)
	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Applications)
}

func TestEventRepositoryGetAllEvents_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	eventService,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventService(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &companyID,
		RecruiterID:          &companyID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, *createApplication1.ID, eventID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Len(t, *(events[0]).Applications, 2)

	assert.Equal(t, application2ID, (*(*events[0]).Applications)[0].ID)

	event1Application2 := (*(*events[0]).Applications)[1]
	assert.Equal(t, *createApplication1.ID, event1Application2.ID)
	assert.Nil(t, event1Application2.CompanyID)
	assert.Nil(t, event1Application2.RecruiterID)
	assert.Nil(t, event1Application2.JobTitle)
	assert.Nil(t, event1Application2.JobAdURL)
	assert.Nil(t, event1Application2.Country)
	assert.Nil(t, event1Application2.Area)
	assert.Nil(t, event1Application2.RemoteStatusType)
	assert.Nil(t, event1Application2.WeekdaysInOffice)
	assert.Nil(t, event1Application2.EstimatedCycleTime)
	assert.Nil(t, event1Application2.EstimatedCommuteTime)
	assert.Nil(t, event1Application2.ApplicationDate)
	assert.Nil(t, event1Application2.CreatedDate)
	assert.Nil(t, event1Application2.UpdatedDate)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplications(t *testing.T) {
	eventService, applicationRepository, companyRepository, eventRepository, _, _, _, _ := setupEventService(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)
	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Applications)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToNone(t *testing.T) {
	eventService,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventService(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	applicationID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, eventID, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, (events[0]).Applications)
}

// -------- GetAllEvents - Company tests: --------

func TestEventRepositoryGetAllEvents_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	eventService, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventService(t)

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID

	// add two companies

	createCompany1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 3, 0))).ID

	// associate events and companies

	Company1Event1 := models.AssociateCompanyEvent{
		CompanyID: *createCompany1.ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company1Event1)
	assert.NoError(t, err)

	Company2Event1 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event1)
	assert.NoError(t, err)

	Company2Event2 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event2)
	assert.NoError(t, err)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 3)

	assert.Equal(t, event1ID, events[0].ID)
	assert.Len(t, *(events[0]).Companies, 2)

	event1Company1 := (*(*events[0]).Companies)[0]
	assert.Equal(t, *createCompany1.ID, event1Company1.ID)
	assert.Equal(t, createCompany1.Name, *event1Company1.Name)
	assert.Equal(t, createCompany1.CompanyType.String(), event1Company1.CompanyType.String())
	assert.Equal(t, createCompany1.Notes, event1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.LastContact, event1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.CreatedDate, event1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.UpdatedDate, event1Company1.UpdatedDate)

	event1Company2 := (*(*events[0]).Companies)[1]
	assert.Equal(t, company2ID, event1Company2.ID)

	assert.Equal(t, event2ID, events[1].ID)
	assert.Len(t, *(events[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(*events[1]).Companies)[0].ID)

	assert.Equal(t, event3ID, events[2].ID)
	assert.Nil(t, events[2].Companies)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventService, _, companyRepository, eventRepository, _, _, _, _ := setupEventService(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Companies)
}

func TestEventRepositoryGetAllEvents_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	eventService, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventService(t)

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID

	// add two companies

	createCompany1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 3, 0))).ID

	// associate events and companies

	Company1Event1 := models.AssociateCompanyEvent{
		CompanyID: *createCompany1.ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company1Event1)
	assert.NoError(t, err)

	Company2Event1 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event1)
	assert.NoError(t, err)

	Company2Event2 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event2)
	assert.NoError(t, err)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 3)

	assert.Equal(t, event1ID, events[0].ID)
	assert.Len(t, *(events[0]).Companies, 2)

	event1Company1 := (*(*events[0]).Companies)[0]
	assert.Equal(t, *createCompany1.ID, event1Company1.ID)
	assert.Nil(t, event1Company1.Name)
	assert.Nil(t, event1Company1.CompanyType)
	assert.Nil(t, event1Company1.Notes)
	assert.Nil(t, event1Company1.LastContact)
	assert.Nil(t, event1Company1.CreatedDate)
	assert.Nil(t, event1Company1.UpdatedDate)

	assert.Equal(t, event2ID, events[1].ID)
	assert.Len(t, *(events[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(*events[1]).Companies)[0].ID)

	assert.Equal(t, event3ID, events[2].ID)
	assert.Nil(t, events[2].Companies)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventService, _, companyRepository, eventRepository, _, _, _, _ := setupEventService(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Companies)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	eventService, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventService(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company and associate it to the event

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Companies)
}

// -------- GetAllEvents - Person tests: --------

func TestEventRepositoryGetAllEvents_ShouldReturnPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	eventService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupEventService(t)

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID

	// add two persons

	var person1Type models.PersonType = models.PersonTypeJobContact
	createPerson1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err := personRepository.Create(&createPerson1)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	// associate events and persons

	Person1Event1 := models.AssociateEventPerson{
		PersonID: *createPerson1.ID,
		EventID:  event1ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person1Event1)
	assert.NoError(t, err)

	Person2Event1 := models.AssociateEventPerson{
		PersonID: person2ID,
		EventID:  event1ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person2Event1)
	assert.NoError(t, err)

	Person2Event2 := models.AssociateEventPerson{
		PersonID: person2ID,
		EventID:  event2ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person2Event2)
	assert.NoError(t, err)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 3)

	assert.Equal(t, event1ID, events[0].ID)
	assert.Len(t, *(events[0]).Persons, 2)

	event1Person1 := (*(*events[0]).Persons)[0]
	assert.Equal(t, *createPerson1.ID, event1Person1.ID)
	assert.Equal(t, createPerson1.Name, *event1Person1.Name)
	assert.Equal(t, createPerson1.PersonType.String(), event1Person1.PersonType.String())
	assert.Equal(t, createPerson1.Email, event1Person1.Email)
	assert.Equal(t, createPerson1.Phone, event1Person1.Phone)
	assert.Equal(t, createPerson1.Notes, event1Person1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createPerson1.CreatedDate, event1Person1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createPerson1.UpdatedDate, event1Person1.UpdatedDate)

	event1Person2 := (*(*events[0]).Persons)[1]
	assert.Equal(t, person2ID, event1Person2.ID)

	assert.Equal(t, event2ID, events[1].ID)
	assert.Len(t, *(events[1]).Persons, 1)
	assert.Equal(t, person2ID, (*(*events[1]).Persons)[0].ID)

	assert.Equal(t, event3ID, events[2].ID)
	assert.Nil(t, events[2].Persons)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToAllAndThereAreNoPersonEventsInRepository(t *testing.T) {
	eventService, _, _, eventRepository, personRepository, _, _, _ := setupEventService(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Persons)
}

func TestEventRepositoryGetAllEvents_ShouldReturnPersonIDsIfIncludePersonsIsSetToIDs(t *testing.T) {
	eventService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupEventService(t)

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID

	// add two persons

	var person1Type models.PersonType = models.PersonTypeJobContact
	createPerson1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err := personRepository.Create(&createPerson1)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	// associate events and persons

	Person1Event1 := models.AssociateEventPerson{
		PersonID: *createPerson1.ID,
		EventID:  event1ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person1Event1)
	assert.NoError(t, err)

	Person2Event1 := models.AssociateEventPerson{
		PersonID: person2ID,
		EventID:  event1ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person2Event1)
	assert.NoError(t, err)

	Person2Event2 := models.AssociateEventPerson{
		PersonID: person2ID,
		EventID:  event2ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&Person2Event2)
	assert.NoError(t, err)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 3)

	assert.Equal(t, event1ID, events[0].ID)
	assert.Len(t, *(events[0]).Persons, 2)

	event1Person1 := (*(*events[0]).Persons)[0]
	assert.Equal(t, *createPerson1.ID, event1Person1.ID)
	assert.Nil(t, event1Person1.Name)
	assert.Nil(t, event1Person1.PersonType)
	assert.Nil(t, event1Person1.Email)
	assert.Nil(t, event1Person1.Phone)
	assert.Nil(t, event1Person1.Notes)
	assert.Nil(t, event1Person1.CreatedDate)
	assert.Nil(t, event1Person1.UpdatedDate)

	assert.Equal(t, event2ID, events[1].ID)
	assert.Len(t, *(events[1]).Persons, 1)
	assert.Equal(t, person2ID, (*(*events[1]).Persons)[0].ID)

	assert.Equal(t, event3ID, events[2].ID)
	assert.Nil(t, events[2].Persons)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToIDsAndThereAreNoPersonEventsInRepository(t *testing.T) {
	eventService, _, _, eventRepository, personRepository, _, _, _ := setupEventService(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Persons)
}

func TestEventRepositoryGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToNone(t *testing.T) {
	eventService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupEventService(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person and associate it to the event

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// get all events

	events, err := eventService.GetAllEvents(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Persons)
}

// -------- Update tests: --------

func TestUpdateEvent_ShouldWork(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	var eventType models.EventType = models.EventTypeCallBooked
	updateEvent := models.UpdateEvent{
		ID:          *createEvent.ID,
		EventType:   &eventType,
		Description: testutil.ToPtr("New Description"),
		Notes:       testutil.ToPtr("New Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, -3, 0)),
	}
	updatedDateApproximation := time.Now()
	err = eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.EventType.String(), event.EventType.String())
	assert.Equal(t, updateEvent.Description, event.Description)
	assert.Equal(t, updateEvent.Notes, event.Notes)
	testutil.AssertEqualFormattedDateTimes(t, updateEvent.EventDate, event.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, event.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, event.UpdatedDate, time.Second)
}

func TestUpdateEvent_ShouldUpdateASingleField(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:        testutil.ToPtr(uuid.New()),
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	updateEvent := models.UpdateEvent{
		ID:    *createEvent.ID,
		Notes: testutil.ToPtr("New Notes"),
	}
	err = eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.Notes, event.Notes)
}

func TestUpdateEvent_ShouldNotReturnErrorIfEventDoesNotExist(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	updateEvent := models.UpdateEvent{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("New Notes"),
	}
	err := eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldDeleteEvent(t *testing.T) {
	eventService, _, _, eventRepository, _, _, _, _ := setupEventService(t)

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	err := eventService.DeleteEvent(&eventID)
	assert.NoError(t, err)

	retrievedPerson, err := eventService.GetEventByID(&eventID)
	assert.Nil(t, retrievedPerson)
	assert.Error(t, err)
}

func TestDeleteEvent_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventService, _, _, _, _, _, _, _ := setupEventService(t)

	id := uuid.New()
	err := eventService.DeleteEvent(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: event does not exist. ID: "+id.String(), notFoundError.Error())
}
