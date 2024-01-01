# Install EMP toolkit
cd libs
echo "Install EMP toolkit"
python3 install_emp.py --deps --tool --ot --sh2pc
cd ..

# Build SMC-Janus and thresholding portion Hyb-Janus
echo "Build SMC-Janus and thresholding portion Hyb-Janus"
cmake -S . -B build
cd build
make -j


