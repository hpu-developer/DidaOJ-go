SELECT cron.schedule('0 3 * * *', $$DELETE FROM didaoj.kv_store WHERE expire_time < now();$$);
