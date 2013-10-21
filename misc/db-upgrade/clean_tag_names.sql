-- Run this script against your TMSU database to fix any invalid tag names.

-- The validation rules for tag names have changed over the version history so
-- it's possible you may have tags with names that are now reserved keywords or
-- contain punctuation which causes problems with the virtual filesystem.

update tag
set name = '_'
where name = '';

-- cannot be '.' or '..'
update tag
set name = '_' || name
where name in ('.' || '..');

-- cannot be logical operator
update tag
set name = '_' || name
where name in ('and', 'or', 'not', 'AND', 'OR', 'NOT');

-- cannot start with '-'
update tag
set name = '_' || substr(name, 2)
where name like '-%';

-- cannot contain '('
update tag
set name = replace(name, '(', '[');

-- cannot contain ')'
update tag
set name = replace(name, ')', ']');

-- cannot contain '='
update tag
set name = replace(name, '=', ':');

-- cannot contain spaces
update tag
set name = replace(name, ' ', '_');

-- cannot contain forward slash
update tag
set name = replace(name, '/', '-');;
