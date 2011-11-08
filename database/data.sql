insert into tag (name) values ('red');
insert into tag (name) values ('square');

insert into file (fingerprint) values('1');
insert into file (fingerprint) values('2');
insert into file (fingerprint) values('3');
insert into file (fingerprint) values('4');
insert into file (fingerprint) values('5');

insert into file_path (file_id, path) values(1, 'apple');
insert into file_path (file_id, path) values(2, 'book');
insert into file_path (file_id, path) values(3, 'kite');
insert into file_path (file_id, path) values(4, 'banana');
insert into file_path (file_id, path) values(5, 'postbox');

insert into file_tag (file_id, tag_id) values(1, 1);
insert into file_tag (file_id, tag_id) values(2, 2);
insert into file_tag (file_id, tag_id) values(3, 1);
insert into file_tag (file_id, tag_id) values(3, 2);
insert into file_tag (file_id, tag_id) values(5, 1);
