
# Releasing

Once your changes are in master, the latest release should be set as a draft at https://github.com/pactflow/terraform/releases/.

Once you've tested that it works as expected:

1. Run `make release` to generate release notes and release commit.
1. Edit the release notes at https://github.com/pactflow/terraform/releases/edit/v<VERSION>.
