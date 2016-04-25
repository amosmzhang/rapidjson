#ifndef __RJ_WRAPPER_H
#define __RJ_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

    typedef void* JsonDoc;
    typedef void* JsonVal;
    JsonDoc JsonInit(void);
    void JsonFree(JsonDoc);
    JsonVal ValInit(void);
    void ValFree(JsonVal);

    void JsonParse(JsonDoc, char *);
    int HasParseError(JsonDoc);

    char *GetString(JsonDoc);

    int HasMember(JsonVal, const char *); 
    int GetMemberCount(JsonVal);
    char *GetMemberName(JsonVal, int);

    JsonVal GetMember(JsonVal, const char *);
    int GetType(JsonVal);
    int IsObj(JsonVal);
    int IsInt(JsonVal);
    int IsDouble(JsonVal);
    int IsBool(JsonVal);
    int IsString(JsonVal);
    int IsArray(JsonVal);
    int IsNull(JsonVal);
    char *ValGetString(JsonVal);
    int ValGetInt(JsonVal);
    double ValGetDouble(JsonVal);
    int ValGetBool(JsonVal);
    char *ValGetBasicString(JsonVal);

    int ValArraySize(JsonVal);
    JsonVal GetArrayValueAt(JsonVal, int);

    void SetInt(JsonVal, int);
    void SetDouble(JsonVal, double);
    void SetString(JsonDoc, JsonVal, const char *);
    void SetBool(JsonVal, int);
    void SetNull(JsonVal);
    void SetValue(JsonVal, JsonVal);
    void InitArray(JsonVal);
    void ArrayAppend(JsonDoc, JsonVal, JsonVal);
    JsonVal InitObj(JsonVal);
    void AddMember(JsonDoc, JsonVal, JsonVal, JsonVal);
    void AddStrMember(JsonDoc, JsonVal, const char *, JsonVal);

    void RemoveMember(JsonVal, const char *);
    void ArrayRemove(JsonVal, int);

#ifdef __cplusplus
}
#endif
#endif
