module github.com/thek4n/DeadmanSwitch

go 1.23.3

require pkg/passphrases v0.0.1
replace pkg/passphrases => ./pkg/passphrases


require github.com/thek4n/DeadmanSwitch/internal/common v0.0.1
replace github.com/thek4n/DeadmanSwitch/internal/common => ./internal/common