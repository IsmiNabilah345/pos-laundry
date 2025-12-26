-- Create users table
CREATE TABLE IF NOT EXISTS public.users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL, -- Note: In a real app, store hashed passwords (e.g., bcrypt)
    role TEXT NOT NULL DEFAULT 'kasir', -- 'admin' or 'kasir'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Insert default admin user if not exists
INSERT INTO public.users (username, password, role)
VALUES ('admin', 'password123', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Insert default kasir user if not exists
INSERT INTO public.users (username, password, role)
VALUES ('kasir', 'kasir123', 'kasir')
ON CONFLICT (username) DO NOTHING;
