#include <cstdlib>
using namespace std;

class BioSetting{
    public:
    string bio_type;   // The type of biometric data in ["finger", "iris"]
    int user_num;      // The number of users in the database
    int fuse;          // The number of templates to be fused for each user
    int template_size; // The size of each template in bits
    int db_size;       // Total number of templates = user_num * fuse
    int threshold;     // The acceptance similarity threshold
    int prime_mod;     // The prime modulus used in the computation
    int bits_per_slot; // The number of bits used to represent a slot in each template scalar

    BioSetting(string type, int N, int f, int ts, int _threshold, int _prime_mod, int _bits_per_slot=8){
        bio_type = type; 
        user_num = N;
        fuse = f;
        template_size = ts;
        db_size = N*f;
        threshold = _threshold;
        prime_mod = _prime_mod;
        if(bio_type == "finger")
            bits_per_slot = _bits_per_slot;
        else if (bio_type == "iris")
            bits_per_slot = 1;
        else
            throw "Unknown biometric type";
    }

    string to_str(){
        return "setting(bio-type: " +  bio_type  + ", TS: "  + to_string(template_size) +  " N: " + to_string(user_num) + " fuse: " + to_string(fuse) + ")";
    }
};

class PlainBiometric{
    public:
    int template_size;
    int bits_per_slot;

    int16_t *data;
    bool *mask;

    PlainBiometric(int sz, int bps=1){
        template_size = sz;
        bits_per_slot = bps;

        data = new int16_t[sz];
        if (bps == 1)
            mask = new bool[sz];
    }
};


int compute_hamming_distance(PlainBiometric *x, PlainBiometric *y){
    int dist = 0;
    for (int i = 0; i < x->template_size; i++){
        if (x->mask[i] && y->mask[i]) 
            dist += x->data[i] ^ y->data[i];
    }
    return dist;
}
int compute_euc_distance(PlainBiometric *x, PlainBiometric *y){
    int dist = 0;
    for (int i = 0; i < x->template_size; i++){
        int diff = x->data[i] - y->data[i];
        dist += diff*diff;
    }
    return dist;
}
int compute_distance(PlainBiometric *x, PlainBiometric *y){
    if (x->bits_per_slot == 1)
        return compute_hamming_distance(x, y);
    else
        return compute_euc_distance(x, y);
}


PlainBiometric* gen_rnd_biometric(int template_size, int bits_per_slot=1, int seed=-1){
    if (seed > 0)
        srand(seed);

    int max_val = (1 << bits_per_slot);
    PlainBiometric *bio = new PlainBiometric(template_size, bits_per_slot);
    for (int i = 0; i < template_size; i++){
        bio->data[i] = rand()%max_val;
        if (bits_per_slot == 1)
            bio->mask[i] = (rand()%10) > 0;
    }
    return bio;
}

PlainBiometric** gen_rnd_bio_db(int db_size, int template_size, int bits_per_slot=1, int seed=-1){
    if (seed > 0)
        srand(seed);

    PlainBiometric **db = new PlainBiometric*[db_size];
    for (int i = 0; i < db_size; i++)
        db[i] = gen_rnd_biometric(template_size, bits_per_slot, -1);
    return db;
}

/**
 * Secret share a biometric template between two parties
 * The sharing is done by generating a random template s1 and compute s2 = x - s1 (mod 2^bits_per_slot)
 * When bits_per_slot = 1, this operation is equevalent to xor and we secret share the mask too
*/
pair<PlainBiometric*, PlainBiometric* > secret_share(PlainBiometric *bio){
    PlainBiometric *s1, *s2 = new PlainBiometric(bio->template_size);
    s1 = gen_rnd_biometric(bio->template_size, bio->bits_per_slot);

    int mod = (1 << bio->bits_per_slot);
    for (int i = 0; i < bio->template_size; i++){
        s2->data[i] = (bio->data[i] - s1->data[i] + mod) % mod;
        if (bio->bits_per_slot == 1){
            s2->mask[i] = bio->mask[i] ^ s1->mask[i];
        } 
    }
    return make_pair(s1, s2);
}


