-- +goose Up
CREATE TABLE services (
                          id    INTEGER PRIMARY KEY,
                          title TEXT NOT NULL
);

INSERT INTO services (id, title) VALUES
                                     (1, 'Третьяковская галерея'),
                                     (2, 'Пушкинский музей'),
                                     (3, 'Театр на Малой Бронной'),
                                     (4, 'Теннис в Лужниках'),
                                     (5, 'Падел корт в Сити'),
                                     (6, 'Дайджест светских событий');

-- +goose Down
DROP TABLE IF EXISTS services;