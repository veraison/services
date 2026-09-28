# test/evidence

`sevsnp-gcp-tsm-report.cbor` (sha256 `ddc3306caf7fd898cf38b41c81cf528aa21959e8cbceb55e0e3c5593d939605a`) is, byte for byte, the TSM report token a
Google Cloud guest (n2d-standard-2, AMD Milan, us-central1-c) made through configfs-tsm on 7 October 2026: the attestation
report signed with the chip's VCEK, and the certificate table as the host supplied it, whose VCEK, ASK and ARK entries
are DER. Its REPORT_DATA is the 64-octet value the guest bound (`sevsnpGCPReportData` in `scheme_test.go`). The
report's trust anchor is AMD's ARK-Milan, as go-sev-guest embeds it.

The other files are the scheme's original samples.
