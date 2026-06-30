# ETL 1 — NetSuite transactions (synthetic)

- **Pipeline / source / target:** netsuite_extract -> transform -> warehouse.fct_transactions
- **Observed failures:** duplicate key on (transaction_id, line_no) in batch 7; recovered after dedup retry.
- **Successful checks:** source vs target row_count diff = 12, fully explained by dropped null account_id rows.
- **Common bugs:** null account_id from NetSuite incremental window; duplicate lines on retried batches.
- **Useful prompts:** "verify target row_count = source row_count minus rows dropped for null account_id".
- **Project knowledge:** dedup on (transaction_id, line_no) is required before load.
- **Data quality risks:** silent row drops if null-account filter is widened without alerting.

> Synthetic example only. Sanitize real client data before committing.
