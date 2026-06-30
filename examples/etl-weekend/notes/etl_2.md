# ETL 2 — Salesforce accounts sync (synthetic)

- **Pipeline / source / target:** sfdc_extract -> transform -> warehouse.dim_account
- **Observed failures:** API timeout on large bulk query; partial batch committed.
- **Successful checks:** primary-key uniqueness on account_id; no orphan child records after reload.
- **Common bugs:** partial commit on bulk API timeout leaves stale rows; idempotent upsert fixes it.
- **Useful prompts:** "check dim_account for duplicate account_id and orphaned child rows after a partial load".
- **Project knowledge:** always upsert by account_id, never insert-only, to survive retries.
- **Data quality risks:** stale attributes if a failed batch is not reprocessed before downstream reads.

> Synthetic example only. Sanitize real client data before committing.
