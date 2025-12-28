-- Create messages table for CDC
CREATE TABLE IF NOT EXISTS public.messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);