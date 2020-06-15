![ubiwhere](https://portugalsmartcities.fil.pt/wp-content/uploads/filexp/147/001.png)
# ubiwhere-challenge
**Author**: Daniel Carvalho

**E-mail**: dacarvalho@ua.pt

# Description
This is my Ubiwhere challenge repository. This program simulates a data acquisition platform working with an external device simulator.
The data created by this external device is produced by a simulator and collected every second.
The percentage of available processor and RAM usage are also collected every second.

A local database was implemented using a Go - Bolt library - which provides tools to create and manage an embedded database.

# 
|Requirement               |Implemented                | Tested |
|----------------|-------------------------------|--------------|
|Collect **CPU** and **RAM**                |:white_check_mark: |:white_check_mark: |
|External device **simulator**          |:white_check_mark: |:white_check_mark: |
|Variables **acquisiton each second**   |:white_check_mark: |:white_check_mark: |
|Get **last n** metrics for **all** variables   |:white_check_mark: |:white_check_mark: |
|Get **last n** metrics for **one or more** variables   |:white_check_mark: |:white_check_mark: |
|Get an **average** of the value of **one or more** variables  |:white_check_mark: |:white_check_mark: |
|Document your application through **comments**  |:white_check_mark: |:white_check_mark: |




> **Note**: The whole program was only tested in **UNIX** environment. Operating system data acquisition commands may not work
correctly in other environments.



