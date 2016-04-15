#include "rapidjson/document.h"
#include "rapidjson/writer.h"
#include "rapidjson/stringbuffer.h"
#include "rjwrapper.h"
#include <iostream>
#include <sstream>

JsonDoc JsonInit() {
    Document *doc = new Document();

    return (void *)doc;
}

void JsonFree(JsonDoc json) {
    Document *doc = (Document *)json;

    delete doc;
}

JsonVal ValInit() {
    Value *val = new Value();

    return (void *)val;
}

void ValFree(JsonVal value) {
    Value *val = (Value *)value;

    delete val;
}

void JsonParse(JsonDoc json, const char *input) {
    ((Document *)json)->Parse(input);
}

int HasParseError(JsonDoc json) {
    return ((Document *)json)->HasParseError();
}

JsonVal GetValue(JsonDoc json) {
    Document *doc = (Document *)json;

    Value *s = doc;
    
    return (void *) s;
}

char *GetString(JsonDoc json) {
    rapidjson::StringBuffer buffer;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
    ((Document *)json)->Accept(writer);
    char *result = strdup(buffer.GetString());

    return result;
}

int HasMember(JsonVal value, const char *member) {
    return ((Value *)value)->HasMember(member);
}

int GetMemberCount(JsonVal value) {
    return ((Value *)value)->MemberCount();
}

char * GetMemberName(JsonVal value, int index) {
    Value::ConstMemberIterator itr = ((Value *)value)->MemberBegin() + index;
    std::string member = itr->name.GetString();

    return strdup(member.c_str());
}

int GetType(JsonVal value) {
    return ((Value *)value)->GetType();
}
int IsObj(JsonVal value) {
    return ((Value *)value)->IsObject();
}
int IsInt(JsonVal value) {
    return ((Value *)value)->IsInt();
}
int IsString(JsonVal value) {
    return ((Value *)value)->IsString();
}
int IsDouble(JsonVal value) {
    return ((Value *)value)->IsDouble();
}
int IsArray(JsonVal value) {
    return ((Value *)value)->IsArray();
}
int IsBool(JsonVal value) {
    return ((Value *)value)->IsBool();
}
int IsNull(JsonVal value) {
    return ((Value *)value)->IsNull();
}

JsonVal GetMember(JsonVal value, const char * key) {
    Value *val = (Value *)value;

    Value& s = (*val)[key];

    return (void *) &s;
}

char *ValGetString(JsonVal value) {
    rapidjson::StringBuffer buffer;
    rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
    ((Value *)value)->Accept(writer);
    char *result = strdup(buffer.GetString());

    return result;
}
int ValGetInt(JsonVal value) {
    return ((Value *)value)->GetInt();
}
double ValGetDouble(JsonVal value) {
    return ((Value *)value)->GetDouble();
}
int ValGetBool(JsonVal value) {
    return ((Value *)value)->GetBool();
}
char * ValGetBasicString(JsonVal value) {
    return strdup( ((Value *)value)->GetString() );
}

int ValArraySize(JsonVal value) {
    return ((Value *)value)->Size();
}
JsonVal GetArrayValueAt(JsonVal value, int index) {
    Value::ConstValueIterator itr = ((Value *)value)->Begin() + index;
    const Value& s = *itr;

    return (void *) &s;
}

void SetInt(JsonVal value, int num) {
    ((Value *)value)->SetInt(num);
}
void SetDouble(JsonVal value, double num) {
    ((Value *)value)->SetDouble(num);
}
void SetString(JsonDoc json, JsonVal value, const char *str) {
    Document *doc = (Document *)json;
    ((Value *)value)->SetString(str, doc->GetAllocator());
}
void SetBool(JsonVal value, int b) {
    ((Value *)value)->SetBool((bool)b);
}
void SetNull(JsonVal value) {
    ((Value *)value)->SetNull();
}
void SetValue(JsonVal value, JsonVal item) {
    *((Value *)value) = *((Value *)item);
}
void InitArray(JsonVal value) {
    ((Value *)value)->SetArray();
}
void ArrayAppend(JsonDoc json, JsonVal value, JsonVal v) {
    Value *val = (Value *)value;
    Value *item = (Value *)v;
    Document *doc = (Document *)json;

    val->PushBack(*item, doc->GetAllocator());
}
JsonVal InitObj(JsonVal value) {
    return (void *) &((Value *)value)->SetObject();
}
void AddMember(JsonDoc json, JsonVal value, const char *k, JsonVal v) {
    Value *val = (Value *)value;
    Value *item = (Value *)v;
    Document *doc = (Document *)json;
    Value key;
    SetString(json, &key, k);

    val->AddMember(key, *item, doc->GetAllocator());
}

void RemoveMember(JsonVal value, const char *k) {
    ((Value *)value)->RemoveMember(k);
}

void ArrayRemove(JsonVal value, int index) {
    Value::ConstValueIterator itr = ((Value *)value)->Begin() + index;
    ((Value *)value)->Erase(itr);
}
