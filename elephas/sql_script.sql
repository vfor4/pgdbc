-- thanks to chatGPT 
CREATE TABLE test_data_types (
    id SERIAL PRIMARY KEY,                -- Auto-incrementing integer
    name VARCHAR(50),                     -- Variable-length string
    age INTEGER,                          -- Integer
    weight FLOAT,                         -- Floating point number
    is_active BOOLEAN,                    -- Boolean
    created_at TIMESTAMP,                 -- Timestamp
    salary NUMERIC(10, 2),                -- Precise numeric type
    dob DATE,                             -- Date
    binary_data BYTEA,                    -- Binary data
    json_data JSON,                       -- JSON data
    jsonb_data JSONB,                     -- JSONB for advanced JSON operations
    ip_address INET,                      -- IP address type
    mac_address MACADDR,                  -- MAC address
    point_data POINT,                     -- Geometric point
    uuid_field UUID DEFAULT gen_random_uuid(), -- Unique identifier
    big_number BIGINT,                    -- Large integer
    interval_field INTERVAL,              -- Time interval
    text_data TEXT,                       -- Large text data
    array_data TEXT[],                    -- Array of text
    enum_status STATUS_TYPE,              -- Enum type (will define it below)
    range_data INT4RANGE                  -- Integer range
);

-- Step 3: Create an ENUM type
CREATE TYPE STATUS_TYPE AS ENUM ('ACTIVE', 'INACTIVE', 'PENDING');

-- Step 4: Insert sample data
INSERT INTO test_data_types (
    name, age, weight, is_active, created_at, salary, dob, binary_data, 
    json_data, jsonb_data, ip_address, mac_address, point_data, big_number, 
    interval_field, text_data, array_data, enum_status, range_data
) VALUES 
('John Doe', 30, 72.5, TRUE, NOW(), 50000.50, '1993-05-20', decode('DEADBEEF', 'hex'), 
 '{"key": "value"}', '{"key": "value", "nested": {"subkey": 123}}', '192.168.1.1', 
 '08:00:2b:01:02:03', POINT(10.5, 20.3), 9223372036854775807, 
 '1 year 2 months', 'This is a test text', ARRAY['one', 'two', 'three'], 'ACTIVE', '[1,10]');
