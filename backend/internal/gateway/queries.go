package gateway

// Job Queries
const (
	GetJobsQuery = `
		query GetJobs($filters: JobFilters, $limit: Int, $offset: Int) {
			jobs(filters: $filters, limit: $limit, offset: $offset) {
				id
				title
				department
				location
				employmentType
				experienceLevel
				salaryRange {
					min
					max
					currency
				}
				description
				requirements
				responsibilities
				benefits
				skills
				status
				postedDate
				closingDate
				applicationCount
				viewCount
				remoteWork
				urgentHiring
				createdBy {
					id
					name
				}
				createdAt
				updatedAt
			}
		}
	`

	GetJobQuery = `
		query GetJob($id: ID!) {
			job(id: $id) {
				id
				title
				department
				location
				employmentType
				experienceLevel
				salaryRange {
					min
					max
					currency
				}
				description
				requirements
				responsibilities
				benefits
				skills
				status
				postedDate
				closingDate
				applicationCount
				viewCount
				remoteWork
				urgentHiring
				createdBy {
					id
					name
					email
				}
				createdAt
				updatedAt
			}
		}
	`

	CreateJobMutation = `
		mutation CreateJob($input: JobInput!) {
			createJob(input: $input) {
				id
				title
				status
				postedDate
			}
		}
	`

	UpdateJobMutation = `
		mutation UpdateJob($id: ID!, $input: JobInput!) {
			updateJob(id: $id, input: $input) {
				id
				title
				status
				updatedAt
			}
		}
	`

	PublishJobMutation = `
		mutation PublishJob($id: ID!) {
			publishJob(id: $id) {
				id
				status
				postedDate
			}
		}
	`

	CloseJobMutation = `
		mutation CloseJob($id: ID!) {
			closeJob(id: $id) {
				id
				status
				closingDate
			}
		}
	`

	DeleteJobMutation = `
		mutation DeleteJob($id: ID!) {
			deleteJob(id: $id)
		}
	`

	IncrementJobViewMutation = `
		mutation IncrementJobView($id: ID!) {
			incrementJobView(id: $id) {
				id
				viewCount
			}
		}
	`
)

// Application Queries
const (
	SubmitApplicationMutation = `
		mutation SubmitApplication($input: ApplicationInput!) {
			submitApplication(input: $input) {
				id
				status
				appliedDate
				aiScore {
					overall
					insights
					strengths
					concerns
					recommendation
					generatedAt
				}
			}
		}
	`

	GetApplicationsQuery = `
		query GetApplications($filters: ApplicationFilters, $limit: Int, $offset: Int) {
			applications(filters: $filters, limit: $limit, offset: $offset) {
				id
				job {
					id
					title
					department
				}
				candidate {
					id
					firstName
					lastName
					email
					phone
					location
				}
				status
				appliedDate
				lastUpdated
				resumeUrl
				coverLetter
				aiScore {
					overall
					recommendation
				}
			}
		}
	`

	GetApplicationQuery = `
		query GetApplication($id: ID!) {
			application(id: $id) {
				id
				job {
					id
					title
					department
					location
					description
					requirements
				}
				candidate {
					id
					firstName
					lastName
					email
					phone
					location
					resumeUrl
					linkedinUrl
					portfolioUrl
				}
				status
				appliedDate
				lastUpdated
				resumeUrl
				coverLetter
				linkedinUrl
				portfolioUrl
				yearsOfExperience
				currentLocation
				willingToRelocate
				expectedSalary
				availability
				aiScore {
					overall
					insights
					strengths
					concerns
					recommendation
					generatedAt
				}
				notes {
					id
					author {
						id
						name
					}
					content
					createdAt
					isInternal
				}
				timeline {
					id
					type
					description
					performedBy {
						id
						name
					}
					timestamp
				}
			}
		}
	`

	UpdateApplicationStatusMutation = `
		mutation UpdateApplicationStatus($id: ID!, $status: ApplicationStatus!, $note: String) {
			updateApplicationStatus(id: $id, status: $status, note: $note) {
				id
				status
				lastUpdated
			}
		}
	`

	BulkUpdateApplicationStatusMutation = `
		mutation BulkUpdateApplicationStatus($ids: [ID!]!, $status: ApplicationStatus!) {
			bulkUpdateApplicationStatus(ids: $ids, status: $status) {
				id
				status
			}
		}
	`

	AddApplicationNoteMutation = `
		mutation AddApplicationNote($applicationId: ID!, $content: String!, $isInternal: Boolean) {
			addApplicationNote(applicationId: $applicationId, content: $content, isInternal: $isInternal) {
				id
				content
				author {
					id
					name
				}
				createdAt
			}
		}
	`

	ScoreApplicationMutation = `
		mutation ScoreApplication($applicationId: ID!) {
			scoreApplication(applicationId: $applicationId) {
				overall
				insights
				strengths
				concerns
				recommendation
				generatedAt
			}
		}
	`
)

// AI Queries
const (
	GenerateJobDescriptionMutation = `
		mutation GenerateJobDescription($input: JobDescriptionInput!) {
			generateJobDescription(input: $input) {
				description
				requirements
				responsibilities
				suggestedSkills
			}
		}
	`
)

// Analytics Queries
const (
	GetRecruitmentMetricsQuery = `
		query GetRecruitmentMetrics($dateRange: DateRangeInput!) {
			recruitmentMetrics(dateRange: $dateRange) {
				totalJobs
				activeJobs
				totalApplications
				avgApplicationsPerJob
				avgTimeToHire
				conversionRates {
					viewToApply
					applyToScreen
					screenToInterview
					interviewToOffer
					offerToAccept
				}
				topPerformingJobs {
					job {
						id
						title
					}
					views
					applications
					conversionRate
					avgTimeToFill
				}
				applicationsByStatus {
					status
					count
					percentage
				}
				applicationTrend {
					date
					value
				}
				sourceBreakdown {
					source
					count
					percentage
				}
			}
		}
	`

	GetJobPerformanceQuery = `
		query GetJobPerformance($jobId: ID!) {
			jobPerformance(jobId: $jobId) {
				job {
					id
					title
				}
				views
				applications
				conversionRate
				avgTimeToFill
			}
		}
	`

	GetApplicationPipelineQuery = `
		query GetApplicationPipeline($jobId: ID) {
			applicationPipeline(jobId: $jobId) {
				status
				count
				applications {
					id
					candidate {
						firstName
						lastName
					}
					appliedDate
					aiScore {
						overall
					}
				}
			}
		}
	`
)

// Candidate Queries
const (
	GetCandidateQuery = `
		query GetCandidate($id: ID!) {
			candidate(id: $id) {
				id
				firstName
				lastName
				email
				phone
				location
				headline
				summary
				resumeUrl
				linkedinUrl
				portfolioUrl
				githubUrl
				skills
				experience {
					company
					title
					startDate
					endDate
					current
					description
					achievements
				}
				education {
					institution
					degree
					field
					startDate
					endDate
					gpa
				}
				certifications {
					name
					issuer
					issueDate
					expiryDate
					credentialId
				}
				languages {
					language
					proficiency
				}
				applications {
					id
					job {
						id
						title
					}
					status
					appliedDate
				}
				availability
				expectedSalary
				preferredLocations
				remotePreference
				createdAt
				updatedAt
			}
		}
	`

	UpdateCandidateProfileMutation = `
		mutation UpdateCandidateProfile($id: ID!, $input: CandidateProfileInput!) {
			updateCandidateProfile(id: $id, input: $input) {
				id
				firstName
				lastName
				updatedAt
			}
		}
	`
)