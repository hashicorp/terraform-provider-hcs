# Copyright IBM Corp. 2020, 2026

schema = 1
artifacts {
  # This should match the `matrix` in .github/workflows/build.yml
  zip = [
    "terraform-provider-hcs_${version}_darwin_amd64.zip",
    "terraform-provider-hcs_${version}_darwin_arm64.zip",
    "terraform-provider-hcs_${version}_freebsd_386.zip",
    "terraform-provider-hcs_${version}_freebsd_amd64.zip",
    "terraform-provider-hcs_${version}_freebsd_arm.zip",
    "terraform-provider-hcs_${version}_linux_386.zip",
    "terraform-provider-hcs_${version}_linux_amd64.zip",
    "terraform-provider-hcs_${version}_linux_arm.zip",
    "terraform-provider-hcs_${version}_linux_arm64.zip",
    "terraform-provider-hcs_${version}_windows_386.zip",
    "terraform-provider-hcs_${version}_windows_amd64.zip",
  ]
}
