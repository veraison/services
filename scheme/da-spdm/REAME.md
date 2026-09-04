# Device Assignment Attestation for SPDM devices

This scheme implements attestation based on EAT-DA tokens described by
[draft-poirier-rats-eat-da]. Currently only SPDM devices are supported (legacy
PCIe devices are not).

> [!NOTE]
> Currently, this implements a PARTIAL attestation based on the EAT-DA token
> alone. The intent is for this to be used as part of composite attestation
> with other schemes (specifically, Arm CCA); as such no integrity validation
> is performed at the moment. The binding to the other scheme's evidence is
> TBD.

[draft-poirier-rats-eat-da]: https://datatracker.ietf.org/doc/draft-poirier-rats-eat-da/
