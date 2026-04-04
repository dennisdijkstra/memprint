ALTER TABLE mem_metadata
  ADD COLUMN IF NOT EXISTS num_goroutines    INTEGER,
  ADD COLUMN IF NOT EXISTS num_cpu           INTEGER,
  ADD COLUMN IF NOT EXISTS go_max_procs      INTEGER,
  ADD COLUMN IF NOT EXISTS num_gc            BIGINT,
  ADD COLUMN IF NOT EXISTS gc_pause_total_ns BIGINT,
  ADD COLUMN IF NOT EXISTS page_size         INTEGER,
  ADD COLUMN IF NOT EXISTS file_pages        INTEGER,
  ADD COLUMN IF NOT EXISTS file_entropy      DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS magic_bytes       TEXT;
