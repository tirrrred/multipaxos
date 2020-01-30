//gRPC tutorial
//https://www.youtube.com/watch?v=Y92WWaZJl24
//https://www.youtube.com/watch?v=_jQ3i_fyqGA&list=PL64wiCrrxh4Jisi7OcCJIUpguV_f5jGnZ&index=15&t=0s

//gRPC uses Protocol buffer. In order to understand gRPC, one need to under
//Protocol buffer as well

//Protocol buffer = data format/structure
//Can be compared with XML, JSON, YAML
//Data kan be serialized and deserialized, what does that mean?
//It is smaller and faster than XML

//.proto file is the definition file for protocol buffer,
//which gives you the data structure
/*
syntax = "proto3";

package todo;

message Task { //"message Task" is creating a type Task with the given data structure:
    text string = 1; //The "=1" is a uniqe identifier for this data entry
	done bool = 2; // "bool" and "string" is data type
	// "text" and "done" are name of the data"
}
*/