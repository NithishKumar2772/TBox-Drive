# TeraBox adapter

The adapter now has an official OAuth client and an initial Open Platform API client for:
- Device QR authorization, code exchange, token info and refresh
- Paginated file listing
- Storage quota lookup

The provider adapter can expose directory listing after it is constructed with a configured API client. Upload, download, folder creation, rename and delete are not yet implemented.

All real account tests require an approved TeraBox Open Platform application and a dedicated authorized test account. Do not use reverse-engineered PCS/NDUS sessions as a substitute for the official OAuth contract.
