package core

import "sync"

// JobStore is an in memory storage interface for jobs
type JobStore struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewJobStore creates a new empty job store
func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

// Add adds a job to the store
func (js *JobStore) Add(j *Job) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.jobs[j.ID] = j
}

// Get returns a job
func (js *JobStore) Get(id string) (j *Job, ok bool) {
	js.mu.RLock()
	defer js.mu.RUnlock()
	j, ok = js.jobs[id]
	return j, ok
}
