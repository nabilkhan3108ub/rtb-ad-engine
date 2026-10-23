-- a handful of sample campaigns across two categories so the
-- auction has real competitors to choose between.
insert into campaigns (name, bid_micros, budget_micros, target_category) values
    ('SportsBrand A',   2500000, 100000000, 'sports'),
    ('SportsBrand B',   3100000, 100000000, 'sports'),
    ('SportsBrand C',   1800000,  50000000, 'sports'),
    ('TechGadget X',    4200000, 200000000, 'tech'),
    ('TechGadget Y',    3900000, 150000000, 'tech'),
    ('TravelDeal Z',    2200000,  80000000, 'travel');
