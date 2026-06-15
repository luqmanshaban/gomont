package core

import "time"


 type URL struct {
    ID int 
    UserID int 
    UserEmail string
    Endpoint string
    IsHealthy bool 
    NotifcationSent bool 
    Retries int
    MaxRetries int 
    Interval int
    RunsAt time.Time
    RetryAt time.Time
    LastManualRetryAt time.Time
    Status string
    CreatedAt time.Time
    UpdatedAt time.Time
 }

 