/**
 * @file    ast_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_exp.h"

char *exp_kinds_[EXP_MAX] = {
    "ID",
    "LIT",
    "TYPE",
    "ARRAY",
    "OP",
    "ACCESS",
    "CALL",
    "SQL",
    "COND",
    "TUPLE"
};

char *op_strs_[OP_MAX] = {
    "ASSIGN",
    "ADD",
    "SUB",
    "MUL",
    "DIV",
    "MOD",
    "AND",
    "OR",
    "BIT_AND",
    "BIT_OR",
    "BIT_XOR",
    "EQ",
    "NE",
    "LT",
    "GT",
    "LE",
    "GE",
    "RSHIFT",
    "LSHIFT",
    "INC",
    "DEC",
    "NOT"
};

char *sql_strs_[SQL_MAX] = {
    "QUERY",
    "INSERT",
    "UPDATE",
    "DELETE"
};

ast_exp_t *
ast_exp_new(exp_kind_t kind, errpos_t *pos)
{
    ast_exp_t *exp = xcalloc(sizeof(ast_exp_t));

    ast_node_init(exp, pos);

    exp->kind = kind;

    ast_meta_init(&exp->meta);

    return exp;
}

ast_exp_t *
exp_null_new(errpos_t *pos)
{
    return ast_exp_new(EXP_NULL, pos);
}

ast_exp_t *
exp_lit_new(errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_LIT, pos);

    ast_val_init(&exp->u_lit.val);

    return exp;
}

ast_exp_t *
exp_type_new(type_t type, char *name, ast_exp_t *k_exp, ast_exp_t *v_exp,
             errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TYPE, pos);

    exp->u_type.type = type;
    exp->u_type.name = name;
    exp->u_type.k_exp = k_exp;
    exp->u_type.v_exp = v_exp;

    return exp;
}

ast_exp_t *
exp_id_new(char *name, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ID, pos);

    exp->u_id.name = name;

    return exp;
}

ast_exp_t *
exp_array_new(ast_exp_t *id_exp, ast_exp_t *idx_exp, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ARRAY, pos);

    exp->u_arr.id_exp = id_exp;
    exp->u_arr.idx_exp = idx_exp;

    return exp;
}

ast_exp_t *
exp_call_new(ast_exp_t *id_exp, array_t *param_exps, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_CALL, pos);

    exp->u_call.id_exp = id_exp;
    exp->u_call.param_exps = param_exps;

    return exp;
}

ast_exp_t *
exp_access_new(ast_exp_t *id_exp, ast_exp_t *fld_exp, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_ACCESS, pos);

    exp->u_acc.id_exp = id_exp;
    exp->u_acc.fld_exp = fld_exp;

    return exp;
}

ast_exp_t *
exp_op_new(op_kind_t kind, ast_exp_t *l_exp, ast_exp_t *r_exp, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_OP, pos);

    exp->u_op.kind = kind;
    exp->u_op.l_exp = l_exp;
    exp->u_op.r_exp = r_exp;

    return exp;
}

ast_exp_t *
exp_ternary_new(ast_exp_t *pre_exp, ast_exp_t *in_exp, ast_exp_t *post_exp,
                errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TERNARY, pos);

    exp->u_tern.pre_exp = pre_exp;
    exp->u_tern.in_exp = in_exp;
    exp->u_tern.post_exp = post_exp;

    return exp;
}

ast_exp_t *
exp_sql_new(sql_kind_t kind, char *sql, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_SQL, pos);

    exp->u_sql.kind = kind;
    exp->u_sql.sql = sql;

    return exp;
}

ast_exp_t *
exp_tuple_new(ast_exp_t *elem_exp, errpos_t *pos)
{
    ast_exp_t *exp = ast_exp_new(EXP_TUPLE, pos);

    exp->u_tup.exps = array_new();

    if (elem_exp != NULL)
        array_add(exp->u_tup.exps, elem_exp);

    return exp;
}

void
ast_exp_dump(ast_exp_t *exp, int indent)
{
}

/* end of ast_exp.c */