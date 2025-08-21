package models

type CompanyType string

const (
	CompanyTypeEmployer    = "employer"
	CompanyTypeRecruiter   = "recruiter"
	CompanyTypeConsultancy = "consultancy"
)

func (companyType CompanyType) IsValid() bool {
	switch companyType {
	case CompanyTypeEmployer, CompanyTypeRecruiter, CompanyTypeConsultancy:
		return true
	}
	return false
}

func (companyType CompanyType) String() string {
	return string(companyType)
}
