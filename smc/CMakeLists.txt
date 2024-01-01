cmake_minimum_required (VERSION 2.8.11)
project (bio-dedup)
set(NAME "bio-dedup")

find_path(CMAKE_FOLDER NAMES cmake/emp-tool-config.cmake)
include(${CMAKE_FOLDER}/cmake/emp-base.cmake)

find_package(emp-ot REQUIRED)
include_directories(${EMP-OT_INCLUDE_DIRS})

# Installation
# install(FILES cmake/emp-sh2pc-config.cmake DESTINATION cmake/)
# install(DIRECTORY emp-sh2pc DESTINATION include/)

find_package(emp-sh2pc REQUIRED)
include_directories(${EMP-SH-2PC_INCLUDE_DIRS})


ENABLE_TESTING()
ADD_SUBDIRECTORY(bio-dedup)
