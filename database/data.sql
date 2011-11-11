insert into tag (name) values ('red');
insert into tag (name) values ('square');

insert into file (path, fingerprint) values('apple', '1');
insert into file (path, fingerprint) values('book', '2');
insert into file (path, fingerprint) values('kite', '3');
insert into file (path, fingerprint) values('banana', '4');
insert into file (path, fingerprint) values('postbox', '5');

insert into file_tag (file_id, tag_id) values(1, 1);
insert into file_tag (file_id, tag_id) values(2, 2);
insert into file_tag (file_id, tag_id) values(3, 1);
insert into file_tag (file_id, tag_id) values(3, 2);
insert into file_tag (file_id, tag_id) values(5, 1);
