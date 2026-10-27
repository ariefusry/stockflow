INSERT INTO products (id, sku, name, price, stock, is_active, created_at) VALUES
(1,  'BRS-PRM-5KG',  'Beras Premium Pandan Wangi 5 kg',        74500, 48,  TRUE,  now() - interval '30 days'),
(2,  'MNY-BMO-2L',   'Minyak Goreng Bimoli 2 L',               38900, 120, TRUE,  now() - interval '30 days'),
(3,  'GLA-GLP-1KG',  'Gula Pasir Gulaku 1 kg',                 17500, 6,   TRUE,  now() - interval '28 days'),
(4,  'MIE-IND-GRG',  'Indomie Goreng Original 85 g',           3500,  400, TRUE,  now() - interval '28 days'),
(5,  'KPI-KPG-165',  'Kopi Kapal Api Special 165 g',           14800, 9,   TRUE,  now() - interval '21 days'),
(6,  'TEH-SSR-25',   'Teh Sosro Celup Isi 25',                 8200,  75,  TRUE,  now() - interval '21 days'),
(7,  'SUS-UHT-1L',   'Susu UHT Ultra Milk Full Cream 1 L',     19500, 0,   TRUE,  now() - interval '20 days'),
(8,  'SBN-LFB-90',   'Sabun Mandi Lifebuoy 90 g',              4500,  210, TRUE,  now() - interval '18 days'),
(9,  'DTR-RNS-800',  'Deterjen Rinso Anti Noda 800 g',         21900, 33,  TRUE,  now() - interval '18 days'),
(10, 'BKT-SDU-38',   'Buku Tulis Sinar Dunia 38 Lembar',       4200,  150, TRUE,  now() - interval '15 days'),
(11, 'PLP-FBC-HTM',  'Pulpen Faber-Castell 1423 Hitam',        3000,  3,   TRUE,  now() - interval '15 days'),
(12, 'TPG-SGB-1KG',  'Tepung Terigu Segitiga Biru 1 kg',       13200, 58,  TRUE,  now() - interval '12 days'),
(13, 'KCP-ABC-135',  'Kecap Manis ABC 135 ml',                 8700,  22,  FALSE, now() - interval '12 days'),
(14, 'AQA-600',      'Air Mineral Aqua 600 ml',                3800,  0,   TRUE,  now() - interval '2 days');

SELECT setval(pg_get_serial_sequence('products', 'id'), (SELECT max(id) FROM products));

INSERT INTO stock_movements (product_id, type, qty, note, created_at) VALUES
(1,  'in',  50,  'stok awal',                      now() - interval '30 days'),
(2,  'in',  126, 'stok awal',                      now() - interval '30 days'),
(3,  'in',  10,  'stok awal',                      now() - interval '28 days'),
(4,  'in',  440, 'stok awal',                      now() - interval '28 days'),
(5,  'in',  9,   'stok awal',                      now() - interval '21 days'),
(6,  'in',  75,  'stok awal',                      now() - interval '21 days'),
(7,  'in',  60,  'kiriman dari distributor',       now() - interval '20 days'),
(7,  'out', 60,  'barang kedaluwarsa ditarik',     now() - interval '5 days'),
(8,  'in',  210, 'stok awal',                      now() - interval '18 days'),
(9,  'in',  33,  'stok awal',                      now() - interval '18 days'),
(10, 'in',  150, 'stok awal',                      now() - interval '15 days'),
(11, 'in',  24,  'stok awal',                      now() - interval '15 days'),
(11, 'out', 21,  'dipakai kantor cabang',          now() - interval '4 days'),
(12, 'in',  58,  'stok awal',                      now() - interval '12 days'),
(13, 'in',  22,  'stok awal',                      now() - interval '12 days');

INSERT INTO sales_orders (id, customer, status, total, created_at, cancelled_at) VALUES
(1, 'Toko Sumber Rejeki Bekasi',   'CREATED',   289000, now() - interval '2 days',  NULL),
(2, 'Warung Bu Siti Yogyakarta',   'CREATED',   303400, now() - interval '1 hour',  NULL),
(3, 'CV Maju Bersama Medan',       'CANCELLED', 74000,  now() - interval '3 days',  now() - interval '3 days' + interval '2 hours');

SELECT setval(pg_get_serial_sequence('sales_orders', 'id'), (SELECT max(id) FROM sales_orders));

INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES
(1, 'BRS-PRM-5KG', 2,  74500),
(1, 'MIE-IND-GRG', 40, 3500),
(2, 'MNY-BMO-2L',  6,  38900),
(2, 'GLA-GLP-1KG', 4,  17500),
(3, 'KPI-KPG-165', 5,  14800);

INSERT INTO stock_movements (product_id, type, qty, note, created_at) VALUES
(1, 'out', 2,  'order #1',        now() - interval '2 days'),
(4, 'out', 40, 'order #1',        now() - interval '2 days'),
(2, 'out', 6,  'order #2',        now() - interval '1 hour'),
(3, 'out', 4,  'order #2',        now() - interval '1 hour'),
(5, 'out', 5,  'order #3',        now() - interval '3 days'),
(5, 'in',  5,  'cancel order #3', now() - interval '3 days' + interval '2 hours');
