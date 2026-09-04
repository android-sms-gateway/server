# android-sms-gateway v1.47.4 Release Notes

## TL;DR

1. Optimized messages index storage to reduce database overhead

## Performance

### Optimized Index Storage

Trimmed the messages user index to reduce storage overhead. This is a database migration that optimizes the index structure for the messages table, resulting in smaller index size and improved query performance.
