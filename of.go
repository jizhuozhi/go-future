package future

import (
	"sync/atomic"
)

func Of2[T0, T1 any](t0 *Future[T0], t1 *Future[T1]) *Future[Tuple2[T0, T1]] {
	var done uint32
	s := &state[Tuple2[T0, T1]]{}
	c := int32(2)

	var res0 T0
	var res1 T1
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple2[T0, T1]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple2[T0, T1]{res0, res1}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple2[T0, T1]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple2[T0, T1]{res0, res1}, nil)
			}
		}
	})

	return &Future[Tuple2[T0, T1]]{state: s}
}

func Of3[T0, T1, T2 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2]) *Future[Tuple3[T0, T1, T2]] {
	var done uint32
	s := &state[Tuple3[T0, T1, T2]]{}
	c := int32(3)

	var res0 T0
	var res1 T1
	var res2 T2
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple3[T0, T1, T2]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple3[T0, T1, T2]{res0, res1, res2}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple3[T0, T1, T2]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple3[T0, T1, T2]{res0, res1, res2}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple3[T0, T1, T2]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple3[T0, T1, T2]{res0, res1, res2}, nil)
			}
		}
	})

	return &Future[Tuple3[T0, T1, T2]]{state: s}
}

func Of4[T0, T1, T2, T3 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3]) *Future[Tuple4[T0, T1, T2, T3]] {
	var done uint32
	s := &state[Tuple4[T0, T1, T2, T3]]{}
	c := int32(4)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple4[T0, T1, T2, T3]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple4[T0, T1, T2, T3]{res0, res1, res2, res3}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple4[T0, T1, T2, T3]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple4[T0, T1, T2, T3]{res0, res1, res2, res3}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple4[T0, T1, T2, T3]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple4[T0, T1, T2, T3]{res0, res1, res2, res3}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple4[T0, T1, T2, T3]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple4[T0, T1, T2, T3]{res0, res1, res2, res3}, nil)
			}
		}
	})

	return &Future[Tuple4[T0, T1, T2, T3]]{state: s}
}

func Of5[T0, T1, T2, T3, T4 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4]) *Future[Tuple5[T0, T1, T2, T3, T4]] {
	var done uint32
	s := &state[Tuple5[T0, T1, T2, T3, T4]]{}
	c := int32(5)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple5[T0, T1, T2, T3, T4]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple5[T0, T1, T2, T3, T4]{res0, res1, res2, res3, res4}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple5[T0, T1, T2, T3, T4]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple5[T0, T1, T2, T3, T4]{res0, res1, res2, res3, res4}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple5[T0, T1, T2, T3, T4]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple5[T0, T1, T2, T3, T4]{res0, res1, res2, res3, res4}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple5[T0, T1, T2, T3, T4]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple5[T0, T1, T2, T3, T4]{res0, res1, res2, res3, res4}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple5[T0, T1, T2, T3, T4]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple5[T0, T1, T2, T3, T4]{res0, res1, res2, res3, res4}, nil)
			}
		}
	})

	return &Future[Tuple5[T0, T1, T2, T3, T4]]{state: s}
}

func Of6[T0, T1, T2, T3, T4, T5 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5]) *Future[Tuple6[T0, T1, T2, T3, T4, T5]] {
	var done uint32
	s := &state[Tuple6[T0, T1, T2, T3, T4, T5]]{}
	c := int32(6)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple6[T0, T1, T2, T3, T4, T5]{res0, res1, res2, res3, res4, res5}, nil)
			}
		}
	})

	return &Future[Tuple6[T0, T1, T2, T3, T4, T5]]{state: s}
}

func Of7[T0, T1, T2, T3, T4, T5, T6 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6]) *Future[Tuple7[T0, T1, T2, T3, T4, T5, T6]] {
	var done uint32
	s := &state[Tuple7[T0, T1, T2, T3, T4, T5, T6]]{}
	c := int32(7)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple7[T0, T1, T2, T3, T4, T5, T6]{res0, res1, res2, res3, res4, res5, res6}, nil)
			}
		}
	})

	return &Future[Tuple7[T0, T1, T2, T3, T4, T5, T6]]{state: s}
}

func Of8[T0, T1, T2, T3, T4, T5, T6, T7 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7]) *Future[Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]] {
	var done uint32
	s := &state[Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]]{}
	c := int32(8)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]{res0, res1, res2, res3, res4, res5, res6, res7}, nil)
			}
		}
	})

	return &Future[Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]]{state: s}
}

func Of9[T0, T1, T2, T3, T4, T5, T6, T7, T8 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8]) *Future[Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]] {
	var done uint32
	s := &state[Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]]{}
	c := int32(9)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]{res0, res1, res2, res3, res4, res5, res6, res7, res8}, nil)
			}
		}
	})

	return &Future[Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]]{state: s}
}

func Of10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9]) *Future[Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]] {
	var done uint32
	s := &state[Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]]{}
	c := int32(10)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9}, nil)
			}
		}
	})

	return &Future[Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]]{state: s}
}

func Of11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10]) *Future[Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]] {
	var done uint32
	s := &state[Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]]{}
	c := int32(11)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10}, nil)
			}
		}
	})

	return &Future[Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]]{state: s}
}

func Of12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10], t11 *Future[T11]) *Future[Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]] {
	var done uint32
	s := &state[Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]]{}
	c := int32(12)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	var res11 T11
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})
	t11.state.subscribe(func(val T11, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{}, err)
			}
		} else {
			res11 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11}, nil)
			}
		}
	})

	return &Future[Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]]{state: s}
}

func Of13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10], t11 *Future[T11], t12 *Future[T12]) *Future[Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]] {
	var done uint32
	s := &state[Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]]{}
	c := int32(13)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	var res11 T11
	var res12 T12
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t11.state.subscribe(func(val T11, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res11 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})
	t12.state.subscribe(func(val T12, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{}, err)
			}
		} else {
			res12 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12}, nil)
			}
		}
	})

	return &Future[Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]]{state: s}
}

func Of14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10], t11 *Future[T11], t12 *Future[T12], t13 *Future[T13]) *Future[Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]] {
	var done uint32
	s := &state[Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]]{}
	c := int32(14)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	var res11 T11
	var res12 T12
	var res13 T13
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t11.state.subscribe(func(val T11, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res11 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t12.state.subscribe(func(val T12, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res12 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})
	t13.state.subscribe(func(val T13, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{}, err)
			}
		} else {
			res13 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13}, nil)
			}
		}
	})

	return &Future[Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]]{state: s}
}

func Of15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10], t11 *Future[T11], t12 *Future[T12], t13 *Future[T13], t14 *Future[T14]) *Future[Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]] {
	var done uint32
	s := &state[Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]]{}
	c := int32(15)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	var res11 T11
	var res12 T12
	var res13 T13
	var res14 T14
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t11.state.subscribe(func(val T11, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res11 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t12.state.subscribe(func(val T12, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res12 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t13.state.subscribe(func(val T13, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res13 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})
	t14.state.subscribe(func(val T14, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{}, err)
			}
		} else {
			res14 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14}, nil)
			}
		}
	})

	return &Future[Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]]{state: s}
}

func Of16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any](t0 *Future[T0], t1 *Future[T1], t2 *Future[T2], t3 *Future[T3], t4 *Future[T4], t5 *Future[T5], t6 *Future[T6], t7 *Future[T7], t8 *Future[T8], t9 *Future[T9], t10 *Future[T10], t11 *Future[T11], t12 *Future[T12], t13 *Future[T13], t14 *Future[T14], t15 *Future[T15]) *Future[Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]] {
	var done uint32
	s := &state[Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]]{}
	c := int32(16)

	var res0 T0
	var res1 T1
	var res2 T2
	var res3 T3
	var res4 T4
	var res5 T5
	var res6 T6
	var res7 T7
	var res8 T8
	var res9 T9
	var res10 T10
	var res11 T11
	var res12 T12
	var res13 T13
	var res14 T14
	var res15 T15
	t0.state.subscribe(func(val T0, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res0 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t1.state.subscribe(func(val T1, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res1 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t2.state.subscribe(func(val T2, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res2 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t3.state.subscribe(func(val T3, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res3 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t4.state.subscribe(func(val T4, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res4 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t5.state.subscribe(func(val T5, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res5 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t6.state.subscribe(func(val T6, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res6 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t7.state.subscribe(func(val T7, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res7 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t8.state.subscribe(func(val T8, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res8 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t9.state.subscribe(func(val T9, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res9 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t10.state.subscribe(func(val T10, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res10 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t11.state.subscribe(func(val T11, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res11 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t12.state.subscribe(func(val T12, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res12 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t13.state.subscribe(func(val T13, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res13 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t14.state.subscribe(func(val T14, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res14 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})
	t15.state.subscribe(func(val T15, err error) {
		if err != nil {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{}, err)
			}
		} else {
			res15 = val
			if atomic.AddInt32(&c, -1) == 0 {
				s.set(Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{res0, res1, res2, res3, res4, res5, res6, res7, res8, res9, res10, res11, res12, res13, res14, res15}, nil)
			}
		}
	})

	return &Future[Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]]{state: s}
}
