package core

import "time"


 type URL struct {
    ID int 
    UserID int 
    Endpoint string
    IsHealthy bool 
    NotifcationSent bool 
    MaxRetries int 
    Interval int
    RunsAt time.Time
    RetryAt time.Time
    LastManualRetryAt time.Time
    LastChecked time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
 }

 