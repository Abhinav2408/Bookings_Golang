#!/bin/bash
#go build is smart to not include test files
go build -o bookings cmd/web/*.go
./bookings -dbname=bookings -dbuser=postgres