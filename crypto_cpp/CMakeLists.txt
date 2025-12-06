cmake_minimum_required(VERSION 3.10)
project(paycrypto)

set(CMAKE_CXX_STANDARD 17)

find_package(OpenSSL REQUIRED)

include_directories(include)

add_library(paycrypto SHARED paycrypto.cpp)
target_link_libraries(paycrypto OpenSSL::SSL OpenSSL::Crypto)

install(TARGETS paycrypto DESTINATION lib)
install(FILES paycrypto.h DESTINATION include)
