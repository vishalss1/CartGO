-- Docker already creates: cartgo_user
-- We only create databases owned by that user

CREATE DATABASE cartgo_user_db OWNER cartgo_user;
CREATE DATABASE cartgo_product_db OWNER cartgo_user;
CREATE DATABASE cartgo_inventory_db OWNER cartgo_user;
CREATE DATABASE cartgo_order_db OWNER cartgo_user;
CREATE DATABASE cartgo_payment_db OWNER cartgo_user;
CREATE DATABASE cartgo_delivery_db OWNER cartgo_user;
CREATE DATABASE cartgo_support_db OWNER cartgo_user;


-- Apply schema + default privileges per database

\connect cartgo_user_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_product_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_inventory_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_order_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_payment_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_delivery_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;

\connect cartgo_support_db
GRANT ALL ON SCHEMA public TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO cartgo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON SEQUENCES TO cartgo_user;