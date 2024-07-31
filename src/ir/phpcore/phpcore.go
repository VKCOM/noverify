package phpcore

import (
	"github.com/VKCOM/noverify/src/ir"
)

func ResolveAlias(function ir.Node) ir.Node {
	nm, ok := function.(*ir.Name)
	if !ok {
		return function
	}
	alias, ok := FuncAliases[nm.Value]
	if ok {
		return alias
	}
	return function
}

func ResolveAliasName(n *ir.Name) *ir.Name {
	alias, ok := FuncAliases[n.Value]
	if ok {
		return alias
	}
	return n
}

var FuncAliases = map[string]*ir.Name{
	// See https://www.php.net/manual/ru/aliases.php

	"chop":                    {Value: "rtrim"},
	"close":                   {Value: "closedir"},
	"com_get":                 {Value: "com_propget"},
	"com_propset":             {Value: "com_propput"},
	"com_set":                 {Value: "com_propput"},
	"die":                     {Value: "exit"},
	"diskfreespace":           {Value: "disk_free_space"},
	"doubleval":               {Value: "floatval"},
	"fputs":                   {Value: "fwrite"},
	"gzputs":                  {Value: "gzwrite"},
	"i18n_convert":            {Value: "mb_convert_encoding"},
	"i18n_discover_encoding":  {Value: "mb_detect_encoding"},
	"i18n_http_input":         {Value: "mb_http_input"},
	"i18n_http_output":        {Value: "mb_http_output"},
	"i18n_internal_encoding":  {Value: "mb_internal_encoding"},
	"i18n_ja_jp_hantozen":     {Value: "mb_convert_kana"},
	"i18n_mime_header_decode": {Value: "mb_decode_mimeheader"},
	"i18n_mime_header_encode": {Value: "mb_encode_mimeheader"},
	"imap_create":             {Value: "imap_createmailbox"},
	"imap_fetchtext":          {Value: "imap_body"},
	"imap_getmailboxes":       {Value: "imap_list_full"},
	"imap_getsubscribed":      {Value: "imap_lsub_full"},
	"imap_header":             {Value: "imap_headerinfo"},
	"imap_listmailbox":        {Value: "imap_list"},
	"imap_listsubscribed":     {Value: "imap_lsub"},
	"imap_rename":             {Value: "imap_renamemailbox"},
	"imap_scan":               {Value: "imap_listscan"},
	"imap_scanmailbox":        {Value: "imap_listscan"},
	"ini_alter":               {Value: "ini_set"},
	"is_double":               {Value: "is_float"},
	"is_integer":              {Value: "is_int"},
	"is_long":                 {Value: "is_int"},
	"is_real":                 {Value: "is_float"},
	"is_writeable":            {Value: "is_writable"},
	"join":                    {Value: "implode"},
	"key_exists":              {Value: "array_key_exists"},
	"ldap_close":              {Value: "ldap_unbind"},
	"mbstrcut":                {Value: "mb_strcut"},
	"mbstrlen":                {Value: "mb_strlen"},
	"mbstrpos":                {Value: "mb_strpos"},
	"mbstrrpos":               {Value: "mb_strrpos"},
	"mbsubstr":                {Value: "mb_substr"},
	"mysql":                   {Value: "mysql_db_query"},
	"mysql_createdb":          {Value: "mysql_create_db"},
	"mysql_db_name":           {Value: "mysql_result"},
	"mysql_dbname":            {Value: "mysql_result"},
	"mysql_dropdb":            {Value: "mysql_drop_db"},
	"mysql_fieldflags":        {Value: "mysql_field_flags"},
	"mysql_fieldlen":          {Value: "mysql_field_len"},
	"mysql_fieldname":         {Value: "mysql_field_name"},
	"mysql_fieldtable":        {Value: "mysql_field_table"},
	"mysql_fieldtype":         {Value: "mysql_field_type"},
	"mysql_freeresult":        {Value: "mysql_free_result"},
	"mysql_listdbs":           {Value: "mysql_list_dbs"},
	"mysql_listfields":        {Value: "mysql_list_fields"},
	"mysql_listtables":        {Value: "mysql_list_tables"},
	"mysql_numfields":         {Value: "mysql_num_fields"},
	"mysql_numrows":           {Value: "mysql_num_rows"},
	"mysql_selectdb":          {Value: "mysql_select_db"},
	"mysql_tablename":         {Value: "mysql_result"},
	"ociassignelem":           {Value: "OCICollection::assignElem"},
	"ocibindbyname":           {Value: "oci_bind_by_name"},
	"ocicancel":               {Value: "oci_cancel"},
	"ocicloselob":             {Value: "OCILob::close"},
	"ocicollappend":           {Value: "OCICollection::append"},
	"ocicollassign":           {Value: "OCICollection::assign"},
	"ocicollmax":              {Value: "OCICollection::max"},
	"ocicollsize":             {Value: "OCICollection::size"},
	"ocicolltrim":             {Value: "OCICollection::trim"},
	"ocicolumnisnull":         {Value: "oci_field_is_null"},
	"ocicolumnname":           {Value: "oci_field_name"},
	"ocicolumnprecision":      {Value: "oci_field_precision"},
	"ocicolumnscale":          {Value: "oci_field_scale"},
	"ocicolumnsize":           {Value: "oci_field_size"},
	"ocicolumntype":           {Value: "oci_field_type"},
	"ocicolumntyperaw":        {Value: "oci_field_type_raw"},
	"ocicommit":               {Value: "oci_commit"},
	"ocidefinebyname":         {Value: "oci_define_by_name"},
	"ocierror":                {Value: "oci_error"},
	"ociexecute":              {Value: "oci_execute"},
	"ocifetch":                {Value: "oci_fetch"},
	"ocifetchinto":            {Value: "oci_fetch_object"},
	"ocifetchstatement":       {Value: "oci_fetch_all"},
	"ocifreecollection":       {Value: "OCICollection::free"},
	"ocifreecursor":           {Value: "oci_free_statement"},
	"ocifreedesc":             {Value: "oci_free_descriptor"},
	"ocifreestatement":        {Value: "oci_free_statement"},
	"ocigetelem":              {Value: "OCICollection::getElem"},
	"ociinternaldebug":        {Value: "oci_internal_debug"},
	"ociloadlob":              {Value: "OCILob::load"},
	"ocilogon":                {Value: "oci_connect"},
	"ocinewcollection":        {Value: "oci_new_collection"},
	"ocinewcursor":            {Value: "oci_new_cursor"},
	"ocinewdescriptor":        {Value: "oci_new_descriptor"},
	"ocinlogon":               {Value: "oci_new_connect"},
	"ocinumcols":              {Value: "oci_num_fields"},
	"ociparse":                {Value: "oci_parse"},
	"ocipasswordchange":       {Value: "oci_password_change"},
	"ociplogon":               {Value: "oci_pconnect"},
	"ociresult":               {Value: "oci_result"},
	"ocirollback":             {Value: "oci_rollback"},
	"ocisavelob":              {Value: "OCILob::save"},
	"ocisavelobfile":          {Value: "OCILob::import"},
	"ociserverversion":        {Value: "oci_server_version"},
	"ocisetprefetch":          {Value: "oci_set_prefetch"},
	"ocistatementtype":        {Value: "oci_statement_type"},
	"ociwritelobtofile":       {Value: "OCILob::export"},
	"ociwritetemporarylob":    {Value: "OCILob::writeTemporary"},
	"odbc_do":                 {Value: "odbc_exec"},
	"odbc_field_precision":    {Value: "odbc_field_len"},
	"pg_clientencoding":       {Value: "pg_client_encoding"},
	"pg_setclientencoding":    {Value: "pg_set_client_encoding"},
	"pos":                     {Value: "current"},
	"recode":                  {Value: "recode_string"},
	"show_source":             {Value: "highlight_file"},
	"sizeof":                  {Value: "count"},
	"snmpwalkoid":             {Value: "snmprealwalk"},
	"strchr":                  {Value: "strstr"},
	"xptr_new_context":        {Value: "xpath_new_context"},
	`checkdnsrr`:              {Value: `dns_check_record`},
	`getmxrr`:                 {Value: `dns_get_mx`},
	`magic_quotes_runtime`:    {Value: `set_magic_quotes_runtime`},
	`stream_register_wrapper`: {Value: `stream_wrapper_register`},
	`set_file_buffer`:         {Value: `stream_set_write_buffer`},
	`socket_set_blocking`:     {Value: `stream_set_blocking`},
	`socket_get_status`:       {Value: `stream_get_meta_data`},
	`socket_set_timeout`:      {Value: `stream_set_timeout`},
}

var TypeToIsFunction = map[string]string{
	"boolean":  "is_bool",
	"integer":  "is_int",
	"double":   "is_float",
	"string":   "is_string",
	"array":    "is_array",
	"object":   "is_object",
	"resource": "is_resource",
}
