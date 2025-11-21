package repositories_test

import (
	"errors"
	"jobsearchtracker/internal/api/v1/requests"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupEventRepository(t *testing.T) (
	*repositories.EventRepository,
	*repositories.ApplicationRepository,
	*repositories.CompanyRepository,
	*repositories.PersonRepository,
	*repositories.ApplicationEventRepository,
	*repositories.CompanyEventRepository,
	*repositories.EventPersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventRepositoryTestContainer(t, *config)

	var eventRepository *repositories.EventRepository
	err := container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
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

	return eventRepository,
		applicationRepository,
		companyRepository,
		personRepository,
		applicationEventRepository,
		companyEventRepository,
		eventPersonRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertEvent(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	insertedEvent, err := eventRepository.Create(&createEvent)
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
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent := models.CreateEvent{
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	createdDateApproximation := time.Now()

	insertedEvent, err := eventRepository.Create(&createEvent)
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

func TestCreate_ShouldReturnConflictErrorOnDuplicateEventID(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	id := uuid.New()

	event1 := models.CreateEvent{
		ID:        &id,
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventRepository.Create(&event1)
	assert.NoError(t, err)

	event2 := models.CreateEvent{
		ID:        &id,
		EventType: models.EventTypeOffer,
		EventDate: time.Now().AddDate(0, 3, 0),
	}
	nilEvent, err := eventRepository.Create(&event2)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetEvent(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 7, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 6, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
	}
	_, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
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

func TestGetByID_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	id := uuid.New()
	nilEvent, err := eventRepository.GetByID(&id)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		notFoundError.Error())
}

// -------- GetAll - Base tests: --------

func TestGetAll_ShouldReturnAllEvents(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)

	createEvent2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	events, err := eventRepository.GetAll(
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

func TestGetAll_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- GetAll - Application tests: --------

func TestEventRepositoryGetAll_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplications(t *testing.T) {
	eventRepository, applicationRepository, companyRepository, _, _, _, _ := setupEventRepository(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)
	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Applications)
}

func TestEventRepositoryGetAll_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplications(t *testing.T) {
	eventRepository, applicationRepository, companyRepository, _, _, _, _ := setupEventRepository(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)
	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Applications)
}

func TestEventRepositoryGetAll_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToNone(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, (events[0]).Applications)
}

// -------- GetAll - Company tests: --------

func TestEventRepositoryGetAll_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	eventRepository, _, companyRepository, _, _, companyEventRepository, _ := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventRepository, _, companyRepository, _, _, _, _ := setupEventRepository(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	eventRepository, _, companyRepository, _, _, companyEventRepository, _ := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventRepository, _, companyRepository, _, _, _, _ := setupEventRepository(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	eventRepository, _, companyRepository, _, _, companyEventRepository, _ := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company and associate it to the event

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// get all events

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Companies)
}

// -------- GetAll - Person tests: --------

func TestEventRepositoryGetAll_ShouldReturnPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, eventPersonRepository := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoPersonsIfIncludePersonsIsSetToAllAndThereAreNoPersonEventsInRepository(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, _ := setupEventRepository(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnPersonIDsIfIncludePersonsIsSetToIDs(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, eventPersonRepository := setupEventRepository(t)

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

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoPersonsIfIncludePersonsIsSetToIDsAndThereAreNoPersonEventsInRepository(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, _ := setupEventRepository(t)

	// setup events

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	events, err := eventRepository.GetAll(
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

func TestEventRepositoryGetAll_ShouldReturnNoPersonsIfIncludePersonsIsSetToNone(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person and associate it to the event

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// get all events

	events, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, events)
	assert.Len(t, events, 1)

	assert.Equal(t, eventID, events[0].ID)
	assert.Nil(t, events[0].Persons)
}

// -------- GetAll - combined objects tests: --------

func TestEventRepositoryGetAll_ShouldReturnTwoEventsEvenIfOneApplicationIsSharedBetweenTwoEvents(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventRepository(t)

	// create two events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create an application and associate it to the events

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		nil).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, event1ID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, event2ID, nil)

	// ensure that two events are returned

	eventsWithApplications, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, eventsWithApplications)
	assert.Len(t, eventsWithApplications, 2)

	assert.Equal(t, event2ID, eventsWithApplications[0].ID)
	assert.Len(t, *eventsWithApplications[0].Applications, 1)

	assert.Equal(t, event1ID, eventsWithApplications[1].ID)
	assert.Len(t, *eventsWithApplications[1].Applications, 1)
}

func TestEventRepositoryGetAll_ShouldReturnTwoEventsEvenIfOneCompanyIsSharedBetweenTwoEvents(t *testing.T) {
	eventRepository, _, companyRepository, _, _, companyEventRepository, _ := setupEventRepository(t)

	// create two events

	event1ID := repositoryhelpers.CreateEvent(t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID

	event2ID := repositoryhelpers.CreateEvent(t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// create a company and associate it to both events

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event2ID, nil)

	// ensure that two companies are returned

	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	assert.Equal(t, event2ID, results[0].ID)
	assert.Len(t, *results[0].Companies, 1)

	assert.Equal(t, event1ID, results[1].ID)
	assert.Len(t, *results[1].Companies, 1)
}

func TestEventRepositoryGetAll_ShouldReturnTwoEventsEvenIfOnePersonIsSharedBetweenTwoEvents(t *testing.T) {
	eventRepository, _, _, personRepository, _, _, eventPersonRepository := setupEventRepository(t)

	// create two events

	event1ID := repositoryhelpers.CreateEvent(t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID

	event2ID := repositoryhelpers.CreateEvent(t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// create a person and associate it to both events

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event1ID, personID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, personID, nil)

	// ensure that two companies are returned

	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	assert.Equal(t, event2ID, results[0].ID)
	assert.Len(t, *results[0].Persons, 1)

	assert.Equal(t, event1ID, results[1].ID)
	assert.Len(t, *results[1].Persons, 1)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithOneApplicationAndTwoCompanies(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		companyEventRepository,
		_ := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two companies and associate them to the event

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// create an application and associate it to the event
	applicationID := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &company1ID, nil, nil).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, eventID, nil)

	// Ensure that the event is returned with one company and two companies
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 1)

	assert.Len(t, *results[0].Companies, 2)
	assert.Equal(t, company2ID, (*results[0].Companies)[0].ID)
	assert.Equal(t, company1ID, (*results[0].Companies)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoApplicationsAndOneCompany(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		companyEventRepository,
		_ := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create a company and associate it to the event
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// create two applications and associate them to the event
	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application1ID, eventID, nil)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// Ensure that the event is returned with one company and two companies

	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 2)
	assert.Equal(t, application1ID, (*results[0].Applications)[0].ID)
	assert.Equal(t, application2ID, (*results[0].Applications)[1].ID)

	assert.Len(t, *results[0].Companies, 1)
	assert.Equal(t, companyID, (*results[0].Companies)[0].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoApplicationsAndTwoCompanies(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		_,
		applicationEventRepository,
		companyEventRepository,
		_ := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two companies and associate them to the event
	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// create two applications and associate them to the event
	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company2ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application1ID, eventID, nil)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company1ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// Ensure that the event is returned with one company and two companies
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 2)
	assert.Equal(t, application1ID, (*results[0].Applications)[0].ID)
	assert.Equal(t, application2ID, (*results[0].Applications)[1].ID)

	assert.Len(t, *results[0].Companies, 2)
	assert.Equal(t, company2ID, (*results[0].Companies)[0].ID)
	assert.Equal(t, company1ID, (*results[0].Companies)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithOneApplicationAndTwoPersons(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		personRepository,
		applicationEventRepository,
		_,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// create an application and associate it to the event
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	applicationID := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, eventID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 1)

	assert.Len(t, *results[0].Persons, 2)
	assert.Equal(t, person2ID, (*results[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*results[0].Persons)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoApplicationsAndOnePerson(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		personRepository,
		applicationEventRepository,
		_,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// create an application and associate it to the event
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application1ID, eventID, nil)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 2)
	assert.Equal(t, application2ID, (*results[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*results[0].Applications)[1].ID)

	assert.Len(t, *results[0].Persons, 1)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoApplicationsAndTwoPersons(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		personRepository,
		applicationEventRepository,
		_,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// create two applications and associate them to the event
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application1ID, eventID, nil)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 2)
	assert.Equal(t, application2ID, (*results[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*results[0].Applications)[1].ID)

	assert.Len(t, *results[0].Persons, 2)
	assert.Equal(t, person2ID, (*results[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*results[0].Persons)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithOneCompanyAndTwoPersons(t *testing.T) {
	eventRepository,
		_,
		companyRepository,
		personRepository,
		_,
		companyEventRepository,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// create a company and associate it to the event
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Companies, 1)

	assert.Len(t, *results[0].Persons, 2)
	assert.Equal(t, person2ID, (*results[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*results[0].Persons)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoCompaniesAndOnePerson(t *testing.T) {
	eventRepository,
		_,
		companyRepository,
		personRepository,
		_,
		companyEventRepository,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// create a company and associate it to the event

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// Ensure that the event is returned with one person and two persons

	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Companies, 2)
	assert.Equal(t, company2ID, (*results[0].Companies)[0].ID)
	assert.Equal(t, company1ID, (*results[0].Companies)[1].ID)

	assert.Len(t, *results[0].Persons, 1)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoCompaniesAndTwoPersons(t *testing.T) {
	eventRepository,
		_,
		companyRepository,
		personRepository,
		_,
		companyEventRepository,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two persons and associate them to the event

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// create two companies and associate it to the event

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Companies, 2)
	assert.Equal(t, company2ID, (*results[0].Companies)[0].ID)
	assert.Equal(t, company1ID, (*results[0].Companies)[1].ID)

	assert.Len(t, *results[0].Persons, 2)
	assert.Equal(t, person2ID, (*results[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*results[0].Persons)[1].ID)
}

func TestEventRepositoryGetAll_ShouldReturnEventWithTwoApplicationsAndTwoCompaniesAndTwoPersons(t *testing.T) {
	eventRepository,
		applicationRepository,
		companyRepository,
		personRepository,
		applicationEventRepository,
		companyEventRepository,
		eventPersonRepository := setupEventRepository(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create two companies and associate it to the event

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// create two applications and associate them to the event
	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company2ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application1ID, eventID, nil)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company2ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// create two persons and associate them to the event

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// Ensure that the event is returned with one person and two persons
	results, err := eventRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	assert.Len(t, *results[0].Applications, 2)
	assert.Equal(t, application1ID, (*results[0].Applications)[0].ID)
	assert.Equal(t, application2ID, (*results[0].Applications)[1].ID)

	assert.Len(t, *results[0].Companies, 2)
	assert.Equal(t, company2ID, (*results[0].Companies)[0].ID)
	assert.Equal(t, company1ID, (*results[0].Companies)[1].ID)

	assert.Len(t, *results[0].Persons, 2)
	assert.Equal(t, person2ID, (*results[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*results[0].Persons)[1].ID)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateEvent(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventRepository.Create(&createEvent)
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
	err = eventRepository.Update(&updateEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
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
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:        testutil.ToPtr(uuid.New()),
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)

	updateEvent := models.UpdateEvent{
		ID:    *createEvent.ID,
		Notes: testutil.ToPtr("New Notes"),
	}
	err = eventRepository.Update(&updateEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.Notes, event.Notes)
}

func TestUpdate_ShouldNotReturnErrorIfEventDoesNotExist(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	updateEvent := models.UpdateEvent{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("New Notes"),
	}
	err := eventRepository.Update(&updateEvent)
	assert.NoError(t, err)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteEvent(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	err := eventRepository.Delete(&eventID)
	assert.NoError(t, err)

	retrievedPerson, err := eventRepository.GetByID(&eventID)
	assert.Nil(t, retrievedPerson)
	assert.Error(t, err)
}

func TestDelete_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventRepository, _, _, _, _, _, _ := setupEventRepository(t)

	id := uuid.New()
	err := eventRepository.Delete(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: event does not exist. ID: "+id.String(), notFoundError.Error())
}
