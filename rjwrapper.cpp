#include "rapidjson/document.h"
#include "rapidjson/writer.h"
#include "rapidjson/stringbuffer.h"
#include "rjwrapper.h"
#include <iostream>
#include <sstream>

JsonDoc JsonInit() {
    rapidjson::Document *doc = new rapidjson::Document();

    return (void *)doc;
}

void JsonFree(JsonDoc json) {
    rapidjson::Document *doc = (rapidjson::Document *)json;

    delete doc;
}

JsonVal ValInit() {
    rapidjson::Value *val = new rapidjson::Value();

    return (void *)val;
}

void ValFree(JsonVal value) {
    rapidjson::Value *val = (rapidjson::Value *)value;

    delete val;
}

void JsonParse(JsonDoc json, const char *input) {
    ((rapidjson::Document *)json)->Parse(input);
}

int HasParseError(JsonDoc json) {
    return ((rapidjson::Document *)json)->HasParseError();
}

JsonVal GetValue(JsonDoc json) {
    rapidjson::Document *doc = (rapidjson::Document *)json;

    rapidjson::Value *s = doc;
    
    return (void *) s;
}

char *GetString(JsonDoc json) {
    rapidjson::StringBuffer buffer;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
    ((rapidjson::Document *)json)->Accept(writer);
    char *result = strdup(buffer.GetString());

    return result;
}

int HasMember(JsonVal value, const char *member) {
    return ((rapidjson::Value *)value)->HasMember(member);
}

int GetMemberCount(JsonVal value) {
    return ((rapidjson::Value *)value)->MemberCount();
}

char * GetMemberName(JsonVal value, int index) {
    rapidjson::Value::ConstMemberIterator itr = ((rapidjson::Value *)value)->MemberBegin() + index;
    std::string member = itr->name.GetString();

    return strdup(member.c_str());
}

int GetType(JsonVal value) {
    return ((rapidjson::Value *)value)->GetType();
}
int IsObj(JsonVal value) {
    return ((rapidjson::Value *)value)->IsObject();
}
int IsInt(JsonVal value) {
    return ((rapidjson::Value *)value)->IsInt();
}
int IsString(JsonVal value) {
    return ((rapidjson::Value *)value)->IsString();
}
int IsDouble(JsonVal value) {
    return ((rapidjson::Value *)value)->IsDouble();
}
int IsArray(JsonVal value) {
    return ((rapidjson::Value *)value)->IsArray();
}
int IsBool(JsonVal value) {
    return ((rapidjson::Value *)value)->IsBool();
}
int IsNull(JsonVal value) {
    return ((rapidjson::Value *)value)->IsNull();
}

JsonVal GetMember(JsonVal value, const char * key) {
    rapidjson::Value *val = (rapidjson::Value *)value;

    rapidjson::Value& s = (*val)[key];

    return (void *) &s;
}

char *ValGetString(JsonVal value) {
    rapidjson::StringBuffer buffer;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
    ((rapidjson::Value *)value)->Accept(writer);
    char *result = strdup(buffer.GetString());

    return result;
}
int ValGetInt(JsonVal value) {
    return ((rapidjson::Value *)value)->GetInt();
}
double ValGetDouble(JsonVal value) {
    return ((rapidjson::Value *)value)->GetDouble();
}
int ValGetBool(JsonVal value) {
    return ((rapidjson::Value *)value)->GetBool();
}
char * ValGetBasicString(JsonVal value) {
    return strdup( ((rapidjson::Value *)value)->GetString() );
}

int ValArraySize(JsonVal value) {
    return ((rapidjson::Value *)value)->Size();
}
JsonVal GetArrayValueAt(JsonVal value, int index) {
    rapidjson::Value::ConstValueIterator itr = ((rapidjson::Value *)value)->Begin() + index;
    const rapidjson::Value& s = *itr;

    return (void *) &s;
}

void SetInt(JsonVal value, int num) {
    ((rapidjson::Value *)value)->SetInt(num);
}
void SetDouble(JsonVal value, double num) {
    ((rapidjson::Value *)value)->SetDouble(num);
}
void SetString(JsonVal value, const char *str) {
    char *s = strdup(str);
    ((rapidjson::Value *)value)->SetString(rapidjson::StringRef(s));
}
void SetBool(JsonVal value, int b) {
    ((rapidjson::Value *)value)->SetBool((bool)b);
}
void SetNull(JsonVal value) {
    ((rapidjson::Value *)value)->SetNull();
}
void SetValue(JsonVal value, JsonVal item) {
    *((rapidjson::Value *)value) = *((rapidjson::Value *)item);
}
void InitArray(JsonVal value) {
    ((rapidjson::Value *)value)->SetArray();
}
void ArrayAppend(JsonDoc json, JsonVal value, JsonVal v) {
    rapidjson::Value *val = (rapidjson::Value *)value;
    rapidjson::Value *item = (rapidjson::Value *)v;
    rapidjson::Document *doc = (rapidjson::Document *)json;

    val->PushBack(*item, doc->GetAllocator());
}
JsonVal InitObj(JsonVal value) {
    return (void *) &((rapidjson::Value *)value)->SetObject();
}
void AddMember(JsonDoc json, JsonVal value, const char *k, JsonVal v) {
    char *key = strdup(k);
    rapidjson::Value *val = (rapidjson::Value *)value;
    rapidjson::Value *item = (rapidjson::Value *)v;
    rapidjson::Document *doc = (rapidjson::Document *)json;

    val->AddMember(rapidjson::StringRef(key), *item, doc->GetAllocator());
}

void RemoveMember(JsonVal value, const char *k) {
    ((rapidjson::Value *)value)->RemoveMember(k);
}

void ArrayRemove(JsonVal value, int index) {
    rapidjson::Value::ConstValueIterator itr = ((rapidjson::Value *)value)->Begin() + index;
    ((rapidjson::Value *)value)->Erase(itr);
}
